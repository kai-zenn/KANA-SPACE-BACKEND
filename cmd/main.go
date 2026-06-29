package main

import (
	"KANA-SPACE-BACKEND/internal/configs"
	"KANA-SPACE-BACKEND/internal/modules/user"
	"KANA-SPACE-BACKEND/internal/pkgs/bcrypt"
	"KANA-SPACE-BACKEND/internal/pkgs/jwt"
	"KANA-SPACE-BACKEND/internal/rest"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

  var expiryTime time.Duration
  if daysStr, found := strings.CutSuffix(conf.JWTExpiry, "d"); found {
  	days, err := strconv.Atoi(daysStr)
  	if err != nil {
  		log.Fatalf("Angka JWT_EXPIRY di .env tidak valid: %v", err)
  	}
  	expiryTime = time.Duration(days) * 24 * time.Hour
	} else {
		expiryTime, err = time.ParseDuration(conf.JWTExpiry)
		if err != nil {
			log.Fatalf("Format JWT_EXPIRY tidak valid: %v", err)
		}
	}
 
 jwtService := jwt.NewJWTToken(conf.JWTSecret, expiryTime)
 // googleVerifier := user.GoogleVerifierInterface(conf.GoogleClientID, nil)

 var bcrypt = bcrypt.NewCryptoBcrypt()
 var storage user.StorageInterface

 router := gin.Default()
 app := rest.NewRest(router, db, jwtService, bcrypt, storage, nil)
 app.MountEndPoint()
 
 fmt.Println("\n  ➜  Local: http://localhost:9090/")
 app.Serve(":9090")
}
