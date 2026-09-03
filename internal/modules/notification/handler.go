package notification

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeNotificationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotificationNotFound), errors.Is(err, ErrInvalidOffset), errors.Is(err, ErrInvalidLimit):
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": err.Error()})
	case errors.Is(err, ErrNotOwner):
		c.JSON(http.StatusForbidden, gin.H{"status": false, "message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": err.Error()})
	}
}

type Handler struct {
	useCase INotificationUseCase
}

func NewNotificationHandler(uc INotificationUseCase) *Handler {
	return &Handler{useCase: uc}
}

func (h *Handler) ListMyNotifications(c *gin.Context) {
	userIDVal, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": false,
			"message": "user_id not found",
		})
		return
	}
	userID, _ := userIDVal.(uuid.UUID)

	onlyUnread := c.Query("only_unread") == "true"

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	res, err := h.useCase.ListMyNotifications(c.Request.Context(), userID, onlyUnread, limit, offset)
	if err != nil {
		writeNotificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Daftar notifikasi berhasil diambil",
		"data":    res,
	})
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	userIDVal, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": false,
			"message": "user_id not found",
		})
		return
	}
	userID, _ := userIDVal.(uuid.UUID)

	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "id notifikasi tidak valid",
		})
		return
	}

	err = h.useCase.MarkAsRead(c.Request.Context(), notifID, userID)
	if err != nil {
		writeNotificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Notifikasi ditandai sebagai sudah dibaca",
	})
}

func (h *Handler) RegisterDevice(c *gin.Context) {
	userIDVal, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": false,
			"message": "user_id not found",
		})
		return
	}
	userID, _ := userIDVal.(uuid.UUID)

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "format request tidak valid: " + err.Error(),
		})
		return
	}

	err := h.useCase.RegisterDevice(c.Request.Context(), userID, req.Token, req.Platform)
	if err != nil {
		writeNotificationError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Device berhasil didaftarkan",
	})
}
