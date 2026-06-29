package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
  useCase IUserUseCase
}

func NewUserHandler(useCase IUserUseCase) *UserHandler {
  return &UserHandler{useCase: useCase}
}

func (h *UserHandler) Register(ctx *gin.Context) {
  var req UserRegisterRequest

  err := ctx.ShouldBindJSON(&req)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status":  false,
      "message": "Format data tidak valid" + err.Error(),
    })
    return
  } 

  if err = h.useCase.Register(ctx.Request.Context(), req); err != nil {
    ctx.JSON(http.StatusUnprocessableEntity, gin.H{
      "status":  false,
      "message": "Gagal mendaftar pengguna" + err.Error(),
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status":  true,
    "message": "Berhasil mendaftar pengguna",
  })
}

func (h *UserHandler) Login(ctx *gin.Context){
  var req UserLoginRequest

  err := ctx.ShouldBindJSON(&req)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Username Wajib diisi",
    })
    return 
  }

  token, err := h.useCase.Login(ctx.Request.Context(), req) 
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal login" + err.Error(),
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status":  true,
    "message": "Berhasil login",
    "token":   token.Token,
  })
}

func (h *UserHandler) LoginWithGoogle(ctx *gin.Context) {
  var req GoogleAuthRequest

  err := ctx.ShouldBindJSON(&req)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status":  false,
      "message": "ID Token Google wajib dikirim" + err.Error(),
    })
    return
  }

  res, err := h.useCase.LoginWithGoogle(ctx.Request.Context(), req)
  if err != nil {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "message": err.Error(),
    })
    return
  } 

  ctx.JSON(http.StatusOK, gin.H{
    "status":  true,
    "message": "Berhasil login",
    "token":   res.Token,
  })
}


func (h *UserHandler) GetProfileByUsername(ctx *gin.Context) {
  username := ctx.Param("username")
  if username == "" {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status":  false,
      "message": "Username Wajib diisi",
    })
    return
  }

  profile, err := h.useCase.GetProfileByUsername(ctx.Request.Context(), username)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status":  false,
      "message": "Profil tidak ditemukan",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status":  true,
    "message": "Berhasil mengambil profil",
    "data":    profile,
  })
}

func (h *UserHandler) UpgradeToSeller(ctx *gin.Context) {
  var req UpgradeSellerRequest

  if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status":  false,
      "message": "Data tidak lengkap" ,
    })
    return
  }

  userID, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status":  false,
      "message": "Sesi anda tidak valid, silahkan login ulang",
    })
    return
  }

  req.UserID = userID.(uuid.UUID)

  err := h.useCase.UpgradeToSeller(ctx.Request.Context(), req)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status":  false,
      "message": "Gagal mengupgrade ke seller" + err.Error(),
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status":  true,
    "message": "Berhasil mengupgrade ke seller",
  })
}

func (h *UserHandler) Update(ctx *gin.Context) {
  var req UpdateProfileRequest

  if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Data tidak lengkap",
    })
  }
}

func (h *UserHandler) UpdateProfile(ctx *gin.Context) {
  var req UpdateProfileRequest

  if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Data tidak lengkap",
    })
    return
  }

  userID, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "message": "Sesi anda tidak valid, silahkan login ulang",
    })
  }

  err := h.useCase.Update(ctx.Request.Context(), userID.(uuid.UUID), req)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status":  false,
      "message": "Gagal memperbarui profil" + err.Error(),
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status":  true,
    "message": "Berhasil memperbarui profil",
  })
}

func (h *UserHandler) UpdatePhotoProfile(ctx *gin.Context) {
  var req PhotoUpdate

  if err := ctx.ShouldBind(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Data tidak valid",
    })
    return
  }

  userID, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "message": "Sesi anda tidak valid, silahkan login ulang",
    })
    return
  }

  req.UserID = userID.(uuid.UUID)
  
  newPhoto, err := h.useCase.UpdatePhotoProfile(req)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status":  false,
      "message": "Gagal memperbarui profil" + err.Error(),
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status":  true,
    "message": "Berhasil memperbarui profil",
    "data": gin.H{
			"profile_photo_link": newPhoto,
		},
  })
}
