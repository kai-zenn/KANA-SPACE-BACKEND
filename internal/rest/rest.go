package rest

import (
	"KANA-SPACE-BACKEND/internal/middlewares"
	"KANA-SPACE-BACKEND/internal/modules/user"
	"KANA-SPACE-BACKEND/internal/pkgs/bcrypt"
	"KANA-SPACE-BACKEND/internal/pkgs/jwt"
	"KANA-SPACE-BACKEND/internal/pkgs/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Rest struct {
  router *gin.Engine
  db     *gorm.DB
  jwtAuth jwt.Interface
  bcrypt  bcrypt.Interface
  storage storage.Interface
  googleVerifier user.GoogleVerifierInterface
}

func NewRest(router *gin.Engine, 
  db *gorm.DB,
  jwtAuth jwt.Interface,
  bcrypt bcrypt.Interface,
  storage storage.Interface,
  googleVerifier user.GoogleVerifierInterface) *Rest {
  return &Rest{
    router: router,
    db: db,
    jwtAuth: jwtAuth,
    bcrypt: bcrypt,
    storage: storage,
    googleVerifier: googleVerifier,
  }
}

func (r *Rest) MountEndPoint() {
  r.router.Static("/uploads", "./uploads")
  
  api := r.router.Group("/api/v1")
  
  userRepo := user.NewUserRepository(r.db)
  userUseCase := user.NewUserUseCase(userRepo, r.bcrypt, r.jwtAuth, r.storage, r.googleVerifier)
  userHandler := user.NewUserHandler(userUseCase)
  
  authGroup := api.Group("/auth")
  {
    authGroup.POST("/register", userHandler.Register)
    authGroup.POST("/login", userHandler.Login)
    authGroup.POST("/google", userHandler.LoginWithGoogle)
  }
  
  userGroup := api.Group("/user")
  userGroup.Use(middlewares.Authenticate(r.jwtAuth))
  {
    userGroup.GET("/profile/:username", userHandler.GetProfileByUsername)
    userGroup.PATCH("/profile", userHandler.UpdateProfile)
    userGroup.PUT("/profile/password", userHandler.UpdatePassword)
    userGroup.POST("/profile/photo", userHandler.UpdatePhotoProfile)
    userGroup.POST("/upgrade", userHandler.UpgradeToSeller)
    userGroup.POST("/:id/follow", userHandler.FollowUsers)
    userGroup.POST("/:id/unfollow", userHandler.UnfollowUser)
  }
} 

func (r *Rest) Serve(port string) {
  r.router.Run(port)
}
