package rest

import (
	"KANA-SPACE-BACKEND/internal/middlewares"
	"KANA-SPACE-BACKEND/internal/modules/space"
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
  nlp space.NLPClientInterface
}

func NewRest(router *gin.Engine, 
  db *gorm.DB,
  jwtAuth jwt.Interface,
  bcrypt bcrypt.Interface,
  storage storage.Interface,
  googleVerifier user.GoogleVerifierInterface,
  nlp space.NLPClientInterface) *Rest {
  return &Rest{
    router: router,
    db: db,
    jwtAuth: jwtAuth,
    bcrypt: bcrypt,
    storage: storage,
    googleVerifier: googleVerifier,
    nlp: nlp,
  }
}

func (r *Rest) MountEndPoint() {
  r.router.Static("/uploads", "./uploads")
  
  api := r.router.Group("/api")

  // -- User Module
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

  // -- Space Module
  spacePostR := space.NewPostRepository(r.db)
  spaceLikeR := space.NewLikeRepository(r.db)
  spaceCommentR := space.NewCommentRepository(r.db)
  
  spacePostUseCase := space.NewPostUseCase(spacePostR, spaceCommentR, spaceLikeR, r.nlp, userRepo, r.storage)
	spaceLikeUseCase := space.NewLikeUseCase(spaceLikeR, spacePostR)
	spaceCommentUseCase := space.NewCommentUseCase(spaceCommentR, spacePostR)

	spaceHandler := space.NewSpaceHandler(spacePostUseCase, spaceCommentUseCase, spaceLikeUseCase)

	spaceGroup := api.Group("/space")
	spaceGroup.Use(middlewares.Authenticate(r.jwtAuth))
	{
		spaceGroup.POST("/posts", spaceHandler.CreatePost)       
		spaceGroup.GET("/posts", spaceHandler.GetFeed)          
		// spaceGroup.GET("/posts/:id", spaceHandler.FindPostByID)  // GET /api/v1/posts/:id (Detail post)
		spaceGroup.DELETE("/posts/:id", spaceHandler.DeletePost)

		spaceGroup.POST("/posts/:id/like", spaceHandler.LikePost)     
		spaceGroup.POST("/posts/:id/unlike", spaceHandler.UnlikePost) 

		spaceGroup.POST("/posts/:id/comments", spaceHandler.CreateComment)    
		spaceGroup.GET("/posts/:id/comments", spaceHandler.GetComments)  
		spaceGroup.DELETE("/posts/comments/:comment_id", spaceHandler.DeleteComment) 
	}
} 

func (r *Rest) Serve(port string) {
  r.router.Run(port)
}
