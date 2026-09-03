package main

import (
	"KANA-SPACE-BACKEND/internal/configs"
	"KANA-SPACE-BACKEND/internal/database"
	"KANA-SPACE-BACKEND/internal/database/seeding"
	"KANA-SPACE-BACKEND/internal/pkgs/bcrypt"
	"KANA-SPACE-BACKEND/internal/pkgs/jwt"
	"KANA-SPACE-BACKEND/internal/pkgs/storage"
	"KANA-SPACE-BACKEND/internal/rest"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
  conf, err := configs.LoadConf()
  if err != nil {
    log.Fatalf("Gagal inisialisasi konfigurasi: %v", err)
  }

  dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
    conf.DBHost, conf.DBUser, conf.DBPassword, conf.DBName, conf.DBPort,
  )
  db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
  if err != nil {
    log.Fatalf("Gagal connect ke database: %v", err)
  }

  if len(os.Args) > 1 {
  	command := os.Args[1]
	  switch command {
			case "migrate":
  			log.Println("Migratin ke database sedang berjalan...")
  			if err := database.Migrate(db); err != nil {
  				log.Fatalf("Gagal migration: %v", err)
  			}
  			log.Println("Migration sukses")
        return

  		case "seed":
  			log.Println("Menyuntikkan seeding...")
  			if err := seeding.SeedDatabase(db); err != nil {
  				log.Fatalf("Gagal seeding: %v", err)
  			}
  			log.Println("Seeding berhasil! Akun Admin siap digunakan.")
  			return
  
  		default:
  			log.Fatalf("Perintah '%s' tidak dikenali. Gunakan 'migrate' atau 'seed'.", command)
		}
  }

  var expiryTime time.Duration
  if daysStr, found := strings.CutSuffix(conf.JWTExpiry, "d"); found {
  	days, err := strconv.Atoi(daysStr)
  	if err != nil {
  		log.Fatalf("Angka JWT_EXPIRY di .env tidak valid: %v ", err)
  	}
  	expiryTime = time.Duration(days) * 24 * time.Hour
	} else {
		expiryTime, err = time.ParseDuration(conf.JWTExpiry)
		if err != nil {
			log.Fatalf("Format JWT_EXPIRY tidak valid: %v ", err)
		}
	}
 
 jwtService := jwt.NewJWTToken(conf.JWTSecret, expiryTime)
 store := storage.NewLocalStorage("uploads/v1/photos")
 // googleVerifier := user.GoogleVerifierInterface(conf.GoogleClientID, nil)

 var bcryptServic = bcrypt.NewCryptoBcrypt()

 firebaseCredPath := conf.FirebaseCredentialsPath
 
 opt := option.WithCredentialsFile(firebaseCredPath)
 firebaseApp, err := firebase.NewApp(context.Background(), nil, opt)
 if err != nil {
	log.Fatalf("gagal inisialisasi firebase app: %v", err)
 }

 router := gin.Default()
 app := rest.NewRest(router, db, jwtService, bcryptServic, store, nil, nil, firebaseApp)
 app.FirebaseApp = firebaseApp 
 app.MountEndPoint()
 
 fmt.Println("\n  ➜  Local: http://localhost:9090/")
 go func() {
  if err := router.Run(":9090"); err != nil && err != http.ErrServerClosed {
      log.Fatalf("Gagal menjalankan server: %v", err)
    }
 }()
 
 quit := make(chan os.Signal, 1)
 signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
 <-quit
 
 log.Println("Shutting down server...")
 
 // Stop scheduler (tunggu job selesai)
 if app.Scheduler != nil {
     ctx := app.Scheduler.Stop()
     <-ctx.Done()
 }
 
 log.Println("Server stopped.")
}
