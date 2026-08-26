package chat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IConversationRepository interface {
	CreateConversation(ctx context.Context, conversation *Conversation) error
	FindByProductAndBuyer(ctx context.Context, productID, buyerID uuid.UUID) (*Conversation, error)
	FindByID(ctx context.Context, conversationId uuid.UUID) (*Conversation, error)
	FindByParticipant(ctx context.Context, userID uuid.UUID) ([]Conversation, error)
}

type IMessageRepository interface {
	CreateMessage(ctx context.Context, message *Message) error
	FindByConversationID(ctx context.Context, conversationID uuid.UUID, cursor time.Time, limit int) ([]Message, error)
	FindByID(ctx context.Context, messageId uuid.UUID) (*Message, error)
	UpdateMessage(ctx context.Context, message *Message) error
}

// == Conversation Repostory ==
type ConversationRepository struct {
  db *gorm.DB
}
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (cr *ConversationRepository) CreateConversation(ctx context.Context, conversation *Conversation) error {
	result := cr.db.WithContext(ctx).Create(conversation)
	return result.Error
}

func (cr *ConversationRepository) FindByProductAndBuyer(ctx context.Context, productID, buyerID uuid.UUID) (*Conversation, error) {
	var conversation Conversation
	err := cr.db.WithContext(ctx).Where("product_id = ? AND buyer_id = ?", productID, buyerID).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (cr *ConversationRepository) FindByID(ctx context.Context, conversationId uuid.UUID) (*Conversation, error) {
	var conversation Conversation
	err := cr.db.WithContext(ctx).Where("id = ?", conversationId).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (cr *ConversationRepository) FindByParticipant(ctx context.Context, userID uuid.UUID) ([]Conversation, error) {
	var conversations []Conversation
	err := cr.db.WithContext(ctx).Where("seller_id = ? OR buyer_id = ?", userID, userID).Find(&conversations).Error
	if err != nil {
		return nil, err
	}
	return conversations, nil 
}

// == Message Repository ==
type MessageRepository struct {
  db *gorm.DB
}
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (mr *MessageRepository) CreateMessage(ctx context.Context, message *Message) error {
	result := mr.db.WithContext(ctx).Create(message)
	return result.Error
}

func (mr *MessageRepository) FindByConversationID(ctx context.Context, conversationID uuid.UUID, cursor time.Time, limit int) ([]Message, error) {
	var messages []Message
	
	db := mr.db.WithContext(ctx).Where("conversation_id ?", conversationID).Order("created_at desc").Limit(limit)
	
	if !cursor.IsZero() {
		db = db.Where("created_at < ?", cursor)
	}
	
	err := db.Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (mr *MessageRepository) FindByID(ctx context.Context, messageId uuid.UUID) (*Message, error) {
	var message Message
	err := mr.db.WithContext(ctx).Where("id = ?", messageId).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (mr *MessageRepository) UpdateMessage(ctx context.Context, message *Message) error {
	result := mr.db.WithContext(ctx).Save(message)
	return result.Error
}
