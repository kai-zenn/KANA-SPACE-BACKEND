package space

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
  postUsecase IPostUseCase
  commentUsecase ICommentUseCase
  likeUsecase ILikeUseCase
}

func NewSpaceHandler(postUsecase IPostUseCase, commentUsecase ICommentUseCase, likeUsecase ILikeUseCase) *Handler {
  return &Handler{
    postUsecase: postUsecase,
    commentUsecase: commentUsecase,
    likeUsecase: likeUsecase,
  }
}

// -- Post Handler / Controller
func (h *Handler) CreatePost(ctx *gin.Context) {
  userIDVal, exists := ctx.Get("user_id")
  if !exists {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status":  false,
      "message": "Sesi tidak valid, silakan login ulang",
    })
    return
  }

  userID, _ := userIDVal.(uuid.UUID)
  // if !ok {
  //   ctx.JSON(http.StatusBadRequest, gin.H{
  //     "status": false,
  //     "message": "User ID tidak valid",
  //   })
  //   return
  // }

  var req CreatePostRequest
  if err := ctx.ShouldBind(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Data tidak lengkap",
    })
    return
  }

  req.UserID = userID

  res, err := h.postUsecase.NewPost(ctx.Request.Context(), req)
  if err != nil {
    if err.Error() == "lokasi wajib diisi untuk tag CariMaterial" {
  		ctx.JSON(http.StatusBadRequest, gin.H{
  			"status":  false,
  			"message": err.Error(),
  		})
  		return
    }
  	ctx.JSON(http.StatusInternalServerError, gin.H{
  		"status":  false,
  		"message": "Gagal membuat postingan",
  	})
  	return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Postingan berhasil dibuat",
    "data": res,
  })
}

func (h *Handler) GetFeed(ctx *gin.Context) {
  userIDVal, exists := ctx.Get("user_id")
  if !exists {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status":  false,
      "message": "Sesi tidak valid, silakan login ulang",
    })
    return
  }

  userID, _ := userIDVal.(uuid.UUID)
  // if !ok {
  //   ctx.JSON(http.StatusBadRequest, gin.H{
  //     "status": false,
  //     "message": "User ID tidak valid",
  //   })
  //   return
  // }
  
  var param FeedQueryParam
  ctx.ShouldBindQuery(&param)

  res, err := h.postUsecase.GetFeed(ctx.Request.Context(), userID, param)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal mengambil feed",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Feed berhasil diambil",
    "posts": res.Posts,
    "next_cursor": res.NextCursor,
  })
}

func (h *Handler) DeletePost(ctx *gin.Context) {
  userIDVal, _ := ctx.Get("user_id")
  roleVal, _ := ctx.Get("role")

  requesterID, _ := userIDVal.(uuid.UUID)
  requesterRole := roleVal.(string)

  postIDstr := ctx.Param("id")
  postID, err := uuid.Parse(postIDstr)
  if err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Post ID tidak valid",
    })
    return
  }

  err = h.postUsecase.DeletePost(ctx.Request.Context(), postID, requesterID, requesterRole)
  if err != nil {
    if err.Error() == "you are not authorized to delete this post" {
      ctx.JSON(http.StatusForbidden, gin.H{
        "status": false,
        "message": err.Error(),
      })
      return
    }
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal menghapus Post",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Post berhasil dihapus",
  })
}

// -- Like Handler / Controller
func (h *Handler) LikePost(ctx *gin.Context) {
  userIDVal, _ := ctx.Get("user_id")
  postIDstr := ctx.Param("id")

  requesterID, _ := userIDVal.(uuid.UUID)
  postID, _ := uuid.Parse(postIDstr)  

  var req LikeParam
  req.UserID = requesterID
  req.PostID = postID
  
  res, err := h.likeUsecase.LikePost(ctx.Request.Context(), req)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal menyukai Post",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Post berhasil disukai",
    "data": res,
  })
}

func (h *Handler) UnlikePost(ctx *gin.Context) {
  userIDVal, _ := ctx.Get("user_id")
  postIDstr := ctx.Param("id")

  userID, _ := userIDVal.(uuid.UUID)
  postID, _ := uuid.Parse(postIDstr)  

  var req LikeParam
  req.PostID = postID
  req.UserID = userID
  
  if err := ctx.ShouldBind(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Invalid request",
    })
    return
  }


  res, err := h.likeUsecase.UnlikePost(ctx.Request.Context(), req)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal membatalkan Like",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "Message": "Like berhasil dibatalkan",
    "data": res,
  })
}


// -- Comment Handler / Controller
func (h *Handler) CreateComment(ctx *gin.Context) {
  userIDVal, _ := ctx.Get("user_id")
  postIDstr := ctx.Param("id")

  userID, _ := userIDVal.(uuid.UUID)
  postID, _ := uuid.Parse(postIDstr)  

  var req CreateCommentRequest
  req.UserID = userID
  req.PostID = postID
  
  if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(http.StatusBadRequest, gin.H{
      "status": false,
      "message": "Invalid request",
    })
    return
  }

  res, err := h.commentUsecase.CreateComment(ctx.Request.Context(), req)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal membuat komentar",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Komentar berhasil dibuat",
    "data": res,
  })
}

func (h *Handler) GetComments(ctx *gin.Context) {
  postIDstr := ctx.Param("id")
  postID, err := uuid.Parse(postIDstr)

  if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
		  "status":  false,
		  "message": "ID postingan tidak valid",
		})
		return
	}

	var req CommentQueryParam
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
		  "status":  false,
		  "message": "Invalid query",
		})
		return
	}

  res, err := h.commentUsecase.GetComments(ctx.Request.Context(), postID, req)
  if err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{
      "status": false,
      "message": "Gagal memuat komentar",
    })
    return
  }

  ctx.JSON(http.StatusOK, gin.H{
    "status": true,
    "message": "Komentar berhasil dimuat",
    "comments": res.Comments,
    "next_cursor": res.NextCursor,
  })
}

func (h *Handler) DeleteComment(ctx *gin.Context) {
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

	commentIDStr := ctx.Param("comment_id") 
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "ID komentar tidak valid",
		})
		return
	}

	err = h.commentUsecase.DeleteComment(ctx.Request.Context(), commentID, requesterID, requestRole)
	if err != nil {
		if err.Error() == "not the owner of this comment" {
			ctx.JSON(http.StatusForbidden, gin.H{
				"status":  false,
				"message": "Anda tidak berhak menghapus komentar ini",
			})
			return
		}

		if err.Error() == "record not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status":  false,
				"message": "Komentar tidak ditemukan",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Gagal menghapus komentar",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Komentar berhasil dihapus",
	})
}
