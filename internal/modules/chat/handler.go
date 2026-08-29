package chat

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeChatError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotParticipant), errors.Is(err, ErrCannotRespondOwnOffer), errors.Is(err, ErrCannotChatOwnProduct):
		c.JSON(http.StatusForbidden, gin.H{"status": false, "message": err.Error()})
	case errors.Is(err, ErrOfferNotFound), errors.Is(err, ErrOfferNotPending),
		errors.Is(err, ErrCannotOfferOnFreeItem), errors.Is(err, ErrOfferPriceRequired):
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
	}
}

type Handler struct {
	chatUseCase    IConversationUseCase
	messageUseCase IMessageUseCase
}

func NewHandler(chatUC IConversationUseCase, messageUC IMessageUseCase) *Handler {
	return &Handler{
		chatUseCase:    chatUC,
		messageUseCase: messageUC,
	}
}

func (h *Handler) GetOrCreateConversation(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)
  
	productID, err := uuid.Parse(ctx.Param("productId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "id produk tidak valid"})
		return
	}

	res, err := h.chatUseCase.GetOrCreateConversation(ctx.Request.Context(), productID, userID)
	if err != nil {
		writeChatError(ctx, err)
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
	"status": true, 
	"message": "Percakapan siap", 
	"data": res,
	})
}

func (h *Handler) ListConversations(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)
	
	res, err := h.chatUseCase.ListConversations(ctx.Request.Context(), userID)
	if err != nil {
		writeChatError(ctx, err)
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
	"status": true, 
	"message": "Daftar percakapan berhasil diambil", 
	"data": res,
	})
}

func (h *Handler) SendMessage(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)
  
	conversationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "id percakapan tidak valid"})
		return
	}

	var req CreateMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
		"status": false, 
		"message": err.Error(),
		})
		return
	}
	req.SenderID = userID

	res, err := h.messageUseCase.SendMessage(ctx.Request.Context(), conversationID, req)
	if err != nil {
		writeChatError(ctx, err)
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
	"status": true, 
	"message": "Pesan terkirim", 
	"data": res,
	})
}

func (h *Handler) GetMessages(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)
  
	conversationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
		"status": false, 
		"message": "id percakapan tidak valid",
		})
		return
	}

	var param MessageListQueryParam
	_ = ctx.ShouldBindQuery(&param)

	res, err := h.messageUseCase.GetMessages(ctx.Request.Context(), conversationID, userID, param)
	if err != nil {
		writeChatError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
	"status": true, 
	"message": "Pesan berhasil diambil", 
	"data": res,
	})
}

func (h *Handler) RespondToOffer(ctx *gin.Context) {
  userIDVal, exist := ctx.Get("user_id")
  if !exist {
    ctx.JSON(http.StatusUnauthorized, gin.H{
      "status": false,
      "error": "user_id not found",
    })
    return
  }
  userID, _ := userIDVal.(uuid.UUID)
  
	messageID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
		"status": false, 
		"message": "id pesan tidak valid",
		})
		return
	}

	var body struct {
		Accept bool `json:"accept"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
		"status": false, 
		"message": err.Error(),
		})
		return
	}

	res, err := h.messageUseCase.RespondToOffer(ctx.Request.Context(), RespondOfferRequest{
		MessageID: messageID, 
		ResponderID: userID, 
		Accept: body.Accept,
	})
	if err != nil {
		writeChatError(ctx, err)
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{
	"status": true, 
	"message": "Offer direspon", 
	"data": res,
	})
}
