package chat

import (
	"context"
	"time"
	"github.com/google/uuid"
)

type IConversationRepository interface {
	CreateConversation(ctx context.Context, c *Conversation) error
	FindByProductAndBuyer(ctx context.Context, productID, buyerID uuid.UUID) (*Conversation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	FindByParticipant(ctx context.Context, userID uuid.UUID) ([]Conversation, error)
}

type IMessageRepository interface {
	CreateMessage(ctx context.Context, m *Message) error
	FindByConversationID(ctx context.Context, conversationID uuid.UUID, cursor time.Time, limit int) ([]Message, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	UpdateMessage(ctx context.Context, m *Message) error
}
