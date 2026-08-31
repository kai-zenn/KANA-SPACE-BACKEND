package lapak

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeProductError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrOnlySellerCanListProduct), errors.Is(err, ErrProductForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrTransactionForbidden),
			errors.Is(err, ErrUserFrozen),
			errors.Is(err, ErrNotOfferOwner):
		c.JSON(http.StatusForbidden, gin.H{"status": false, "message": err.Error()})
	case errors.Is(err, ErrInvalidPriceRange),
		  errors.Is(err, ErrListingTypeMismatch),
			errors.Is(err, ErrCategoryMustBeSubcategory),
			errors.Is(err, ErrProductLocked),
			errors.Is(err, ErrOfferRejected),
			errors.Is(err, ErrOfferNotAccepted),
			errors.Is(err, ErrProductNotAvailable):
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

type handler struct {
  productUsecase IProductUseCase
  categoryUsecase ICategoryUseCase
  transactionUsecase ITransactionUseCase
}

func NewLapakHandler(productUsecase IProductUseCase, categoryUsecase ICategoryUseCase, transactionUsecase ITransactionUseCase) *handler {
  return &handler{
    productUsecase: productUsecase,
    categoryUsecase: categoryUsecase,
    transactionUsecase: transactionUsecase,
  }
}

// -- Produk Handler / Controller
func (h *handler) CreateProduct(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)
  // if !ok {
  //   ctx.JSON(http.StatusUnauthorized, gin.H{
  //     "status": false,
  //     "error": "user_id not found",
  //   })
  //   return
  // }

  var req CreateProductRequest
  if err := ctx.ShouldBind(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": err.Error(),
    })
    return
  }

  req.UserID = userID
  res, err := h.productUsecase.NewProduct(ctx.Request.Context(), req, "seller")
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": err.Error(),
    })
    return
  }
  
  ctx.JSON(http.StatusCreated, gin.H{
    "status": true,
    "message": "Produk berhasil ditambahkan",
    "data": res,
  })
}

func (h *handler) GetProductList(ctx *gin.Context) {
  // userIDVal, exist := ctx.Get("user_id")
  // if !exist {
  //   ctx.JSON(http.StatusUnauthorized, gin.H{
  //     "status": false,
  //     "error": "Sesi tidak valid, silakan login ulang",
  //   })
  //   return
  // }
  // userID, _ := userIDVal.(uuid.UUID)

  var param ProductListQueryParam
  if err := ctx.ShouldBindQuery(&param); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Gagal mengambil Produk" + err.Error(),
    })
    return
  }

  res, err := h.productUsecase.GetProductList(ctx.Request.Context(), param)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Gagal mengambil Produk",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Produk berhasil diambil",
    "products": res.Products,
    "next_cursor": res.NextCursor,
  }) 
}

func (h *handler) GetProductByID(ctx *gin.Context) {
  productIDstr := ctx.Param("id")
  productID, err := uuid.Parse(productIDstr)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Product ID tidak valid",
    })
    return
  }

  res, err := h.productUsecase.FindProductByID(ctx.Request.Context(), productID)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal mengambil Produk",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Produk berhasil diambil",
    "data": res,
  })
}

func (h *handler) UpdateProduct(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "Sesi tidak valid, silakan login ulang",
    })
    return
  }
  requesterID, _ := userIDVal.(uuid.UUID)

  var req UpdateProductRequest
  if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Data tidak lengkap",
    })
    return
  }
  req.UserID = requesterID

  productIDstr := ctx.Param("id")
  productID, err := uuid.Parse(productIDstr)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Product ID tidak valid",
    })
    return
  }
  req.ProductID = productID

  res, err := h.productUsecase.UpdateProduct(ctx.Request.Context(), req)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal memperbarui Produk",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Produk berhasil diperbarui",
    "data": res,
  })
}

func (h *handler) DeleteProduct(ctx *gin.Context) {
 	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": "Sesi tidak valid, silakan login ulang",
		})
		return
	}
 
	var requesterID uuid.UUID
	if strID, ok := userIDVal.(string); ok {
		requesterID, _ = uuid.Parse(strID)
	} else if uuidID, ok := userIDVal.(uuid.UUID); ok {
		requesterID = uuidID
	}

	roleVal, _ := ctx.Get("role")
	requestRole, _ := roleVal.(string)
	
  productIDstr := ctx.Param("id")
  productID, err := uuid.Parse(productIDstr)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Product ID tidak valid",
    })
    return
  }

  err = h.productUsecase.DeleteProduct(ctx.Request.Context(), productID, requesterID, requestRole)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal menghapus Produk",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Produk berhasil dihapus",
  })
}

// -- Category Handler / Controller
func (h *handler) GetCategories(ctx *gin.Context) {
	categories, err := h.categoryUsecase.GetCategoryTree(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": false,
			"message": "Gagal mengambil kategori",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": true,
		"message": "Kategori berhasil diambil",
		"categories": categories,
	})
}

// -- Transaction Handler / Controller
func (h *handler) CreateTransaction(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)

  productID, err := uuid.Parse(ctx.Param("id"))
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Product ID tidak valid",
    })
    return
  }

  // ini untuk mendapatkan koordinat pembeli dari request, mobile harus mengirimkan koordinat pembeli dari GPS
  // keknya bakal gw optimalkan lagi kedepannya
 	var body struct {
		Quantity  int     `json:"quantity" binding:"required,min=1"`
		Latitude  float64 `json:"buyer_latitude" binding:"required"`
		Longitude float64 `json:"buyer_longitude" binding:"required"`
	}
	
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "koordinat pembeli wajib dikirim",
			"error":   err.Error(),
		})
		return
	}

	
	req := CreateTransactionRequest{
		ProductID: productID,
		BuyerID:   userID,
		Quantity:  body.Quantity,
	}
	
	res, err := h.transactionUsecase.NewTransaction(ctx.Request.Context(), req, body.Latitude, body.Longitude)
	if err != nil {
		writeProductError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": true,
		"message": "Transaksi berhasil dibuat",
		"data": res,
	})
}

func (h *handler) ConfirmTransaction(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)

  txID, err := uuid.Parse(ctx.Param("id"))
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Transaksi tidak valid",
    })
    return
  }

  err = h.transactionUsecase.ConfirmTransaction(ctx.Request.Context(), txID, userID)
  if err != nil {
    writeProductError(ctx, err)
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Transaksi berhasil dikonfirmasi",
  })
}

func (h *handler) ConfirmViaQR(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)

  var body struct {
    QRCode string `json:"qr_code" binding:"required"`
  }

  if err := ctx.ShouldBindJSON(&body); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": err.Error(),
    })
    return
  }

  err := h.transactionUsecase.CompleteViaQR(ctx.Request.Context(), body.QRCode, userID)
  if err != nil {
    writeProductError(ctx, err)
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Transaksi berhasil dikonfirmasi",
  })
}

func (h *handler) CompleteManual(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)

  txID, err := uuid.Parse(ctx.Param("id"))
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Transaksi tidak valid",
    })
    return
  }

  err = h.transactionUsecase.CompleteManual(ctx.Request.Context(), txID, userID)
  if err != nil {
    writeProductError(ctx, err)
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Transaksi berhasil dikonfirmasi",
  })
}

func (h *handler) CancelTransaction(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)

  txID, err := uuid.Parse(ctx.Param("id"))
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Transaksi tidak valid",
    })
    return
  }

  err = h.transactionUsecase.CancelTransaction(ctx.Request.Context(), txID, userID)
  if err != nil {
    writeProductError(ctx, err)
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Transaksi berhasil dibatalkan",
  })
}

func (h *handler) CheckoutFromOffer(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)

	var req CheckoutFromOfferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
		return
	}

	res, err := h.transactionUsecase.CheckoutFromOffer(ctx.Request.Context(), req.OfferID, userID, req.BuyerLat, req.BuyerLng)
	if err != nil {
		writeProductError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
	"status": true, 
	"message": "Checkout berhasil, transaksi dibuat", 
	"data": res,
	})
}
