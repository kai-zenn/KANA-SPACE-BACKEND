package chat
import (
	"time"
	"github.com/google/uuid"
)

const (
	MessageTypeText  = "TEXT"
	MessageTypeOffer = "OFFER"
)

const (
	OfferStatusPending  = "PENDING"
	OfferStatusAccepted = "ACCEPTED"
	OfferStatusRejected = "REJECTED"
)

type Conversation struct {
	ID            uuid.UUID `gorm:"primary_key"`
	TransactionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"` // 1 transaksi = 1 percakapan
	ProductID     uuid.UUID `gorm:"type:uuid;not null;index"`
	SellerID      uuid.UUID `gorm:"type:uuid;not null;index"`
	BuyerID       uuid.UUID `gorm:"type:uuid;not null;index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Message struct {
	ID             uuid.UUID `gorm:"primary_key"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;index"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null"`

	Type    string `gorm:"type:varchar(20);not null;default:'TEXT'"`
	Content string `gorm:"type:text;not null"`

	OfferPrice  *int
	OfferStatus *string `gorm:"type:varchar(20)"`

	CreatedAt time.Time
}
