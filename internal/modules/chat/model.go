package chat

import (
	"time"

	"github.com/google/uuid"
)

type ConversationResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	SellerID  uuid.UUID `json:"seller_id"`
	BuyerID   uuid.UUID `json:"buyer_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversationListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
}

type CreateMessageRequest struct {
	ConversationID uuid.UUID `json:"-"`
	SenderID       uuid.UUID `json:"-"`
	Type           string    `json:"type" binding:"required,oneof=TEXT OFFER"`
	Content        string    `json:"content" binding:"required"`
	OfferPrice     *int      `json:"offer_price"`
}

type MessageResponse struct {
	ID          uuid.UUID `json:"id"`
	SenderID    uuid.UUID `json:"sender_id"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	OfferPrice  *int      `json:"offer_price,omitempty"`
	OfferStatus *string   `json:"offer_status,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type MessageListQueryParam struct {
	Cursor string `form:"cursor"`
	Limit  int    `form:"limit"`
}

type MessageListResponse struct {
	Messages   []MessageResponse `json:"messages"`
	NextCursor *time.Time        `json:"next_cursor,omitempty"`
}

type RespondOfferRequest struct {
	MessageID   uuid.UUID `json:"-"`
	ResponderID uuid.UUID `json:"-"`
	Accept      bool      `json:"accept"`
}
