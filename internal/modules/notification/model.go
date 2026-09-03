package notification

import (
	"time"

	"github.com/google/uuid"
)

type NotifyInput struct {
	UserID        uuid.UUID
	Type          string
	Title         string
	Body          string
	ReferenceType string 
	ReferenceID   *uuid.UUID
	Data          map[string]string
}

type NotificationResponse struct {
	ID            uuid.UUID  `json:"id"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	Body          string     `json:"body"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	IsRead        bool       `json:"is_read"`
	CreatedAt     time.Time  `json:"created_at"`
}

type RegisterDeviceRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform" binding:"required,oneof=android ios web"`
}
