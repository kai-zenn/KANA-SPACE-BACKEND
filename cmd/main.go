package cmd

import (
	"KANA-SPACE-BACKEND/internal/configs"
	"KANA-SPACE-BACKEND/internal/modules/user"
	"KANA-SPACE-BACKEND/internal/pkgs/jwt"
	"KANA-SPACE-BACKEND/internal/rest"
	"fmt"
	"log"
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
 
 
 jwtService := jwt.NewJWTToken(conf.JWTSecret, time.Duration(conf.JWTExpiry))
 // googleVerifier := user.GoogleVerifierInterface(conf.GoogleClientID, nil)

 var bcrypt user.BcryptInterface
 var storage user.StorageInterface

 router := gin.Default()
 app := rest.NewRest(router, db, jwtService, bcrypt, storage, nil)
 app.MountEndPoint()
 app.Serve(":9090")
}
