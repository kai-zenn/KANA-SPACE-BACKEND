package rest

import (
	"KANA-SPACE-BACKEND/internal/adapters"
	"KANA-SPACE-BACKEND/internal/middlewares"
	"KANA-SPACE-BACKEND/internal/modules/chat"
	"KANA-SPACE-BACKEND/internal/modules/lapak"
	"KANA-SPACE-BACKEND/internal/modules/space"
	"KANA-SPACE-BACKEND/internal/modules/user"
	"KANA-SPACE-BACKEND/internal/pkgs/bcrypt"
	"KANA-SPACE-BACKEND/internal/pkgs/jwt"
	"KANA-SPACE-BACKEND/internal/pkgs/storage"
	"KANA-SPACE-BACKEND/internal/workers"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
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
  Scheduler *cron.Cron
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

	// -- Lapak Module
	lapakProductR := lapak.NewProductRepository(r.db)
	lapakCategoryR := lapak.NewCategoryRepository(r.db)
	lapakTransactionR := lapak.NewTransactionRepository(r.db)

	chatRepo := chat.NewConversationRepository(r.db)
	messageRepo := chat.NewMessageRepository(r.db)

	chatAdapter := adapters.NewChatAdapter(messageRepo, chatRepo)
	
	productUseCase := lapak.NewProductUseCase(lapakProductR, lapakCategoryR, userRepo, r.nlp, r.storage)
	categoryUseCase := lapak.NewCategoryUseCase(lapakCategoryR)
	transactionUseCase := lapak.NewTransactionUseCase(lapakTransactionR, lapakProductR, userRepo, chatAdapter)
	
	autoCancelWorker := workers.NewAutoCancelWorker(transactionUseCase)
	scheduler := workers.NewScheduler(autoCancelWorker)
	scheduler.Start()
	r.Scheduler = scheduler
	
	lapakHandler := lapak.NewLapakHandler(productUseCase, categoryUseCase, transactionUseCase)

	lapakGroup := api.Group("/lapak")

	// Public endpoints
	lapakGroup.GET("/products", lapakHandler.GetProductList)
	lapakGroup.GET("products/:id", lapakHandler.GetProductByID)
	lapakGroup.GET("/categories", lapakHandler.GetCategories)
	lapakGroup.GET("/products/nearby", lapakHandler.GetProductsNearby)

	// Protected / Authenticated endpoints
	lapakGroup.Use(middlewares.Authenticate(r.jwtAuth))
	{
	  // Product endpoints
		lapakGroup.POST("products", lapakHandler.CreateProduct)
		lapakGroup.PATCH("products/:id", lapakHandler.UpdateProduct)
		lapakGroup.DELETE("products/:id", lapakHandler.DeleteProduct)

		// Transaction endpoints
		lapakGroup.POST("/products/:id/transactions", lapakHandler.CreateTransaction)
		lapakGroup.POST("/transactions/:id/confirm", lapakHandler.ConfirmTransaction)   // FINISHED_GOODS
		lapakGroup.POST("/transactions/complete-qr", lapakHandler.ConfirmViaQR)         // RAW_MATERIAL
		lapakGroup.POST("/transactions/:id/complete", lapakHandler.CompleteManual)       // FINISHED_GOODS
		lapakGroup.POST("/transactions/:id/cancel", lapakHandler.CancelTransaction)  

		// Checkout dari offer
		lapakGroup.POST("/transactions/checkout-offer", lapakHandler.CheckoutFromOffer)
	}

	// -- Chat & Nego Module
	lapakAdapter := adapters.NewLapakAdapter(lapakProductR)
	
	chatUseCase := chat.NewConversationUseCase(chatRepo, lapakAdapter)
	messageUseCase := chat.NewMessageUseCase(chatRepo, messageRepo, lapakAdapter)
	
	chatHandler := chat.NewHandler(chatUseCase, messageUseCase)
	
	chatGroup := api.Group("/chat")
		  
	chatGroup.Use(middlewares.Authenticate(r.jwtAuth))
	{
	  // Endpoint terkait Percakapan (Conversation)
		chatGroup.GET("/conversations", chatHandler.ListConversations)
		chatGroup.POST("/products/:productId/conversation", chatHandler.GetOrCreateConversation)
		
		// Endpoint terkait Pesan & Offer (Message)
		chatGroup.POST("/conversations/:id/messages", chatHandler.SendMessage)
		chatGroup.GET("/conversations/:id/messages", chatHandler.GetMessages)

		// Endpoint khusus merespon Offer (Accept/Reject)
		chatGroup.POST("/messages/:id/respond-offer", chatHandler.RespondToOffer)
	}
} 

func (r *Rest) Serve(port string) {
  r.router.Run(port)
}
