package lapak


import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"KANA-SPACE-BACKEND/internal/modules/user"
)

const (
	CategoryBranchRawMaterial   = "RAW_MATERIAL"
	CategoryBranchFinishedGoods = "FINISHED_GOODS"
)

const (
	ListingTypeHibah        = "HIBAH"
	ListingTypeJualBorongan = "JUAL_BORONGAN"
	ListingTypeDijual       = "DIJUAL"
)

const (
	ProductStatusAvailable = "AVAILABLE"
	ProductStatusInactive  = "INACTIVE"
	ProductStatusLocked    = "LOCKED"
	ProductStatusCompleted = "COMPLETED"
)

const (
	TransactionStatusLocked    = "LOCKED"     
	TransactionStatusPending   = "PENDING"   
	TransactionStatusConfirmed = "CONFIRMED"
	TransactionStatusCompleted = "COMPLETED"
	TransactionStatusExpired   = "EXPIRED" 
	TransactionStatusCancelled = "CANCELLED" 
)

const (
	LogisticTypeSelfPickup      = "SELF_PICKUP"
	LogisticType3PL             = "3PL"
	LogisticTypeArrangedOffline = "ARRANGED_OFFLINE"
)

type Category struct {
	ID       uuid.UUID  `gorm:"primary_key"`
	Name     string     `gorm:"type:varchar(100);not null"`
	Slug     string     `gorm:"type:varchar(100);not null;unique"`
	Branch   string     `gorm:"type:varchar(20);not null;index"`
	ParentID *uuid.UUID `gorm:"type:uuid;index"`
}

type Product struct {
	ID     uuid.UUID `gorm:"primary_key"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`

	Title       string `gorm:"type:varchar(150);not null"`
	Description string `gorm:"type:text;not null"`

	CategoryID uuid.UUID `gorm:"type:uuid;not null;index"`
	Category   Category  `gorm:"foreignKey:CategoryID"`

	ListingType        string  `gorm:"type:varchar(20);not null"`
	Price              int     `gorm:"not null;default:0"`
	SelfDeclarationTag *string `gorm:"type:varchar(20)"`
	Status             string  `gorm:"type:varchar(20);not null;default:'AVAILABLE'"`

	Latitude  float64 `gorm:"not null"`
	Longitude float64 `gorm:"not null"`

	Images []ProductImage `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`

	Embedding      pgtype.FlatArray[float64] `gorm:"type:float8[]"`
	EmbeddingModel string                    `gorm:"type:varchar(50)"`

	User user.User `gorm:"foreignKey:UserID;references:ID"`

	Stock *int `gorm:"default:null"` // nil = RAW_MATERIAL (nggak relevan), terisi = FINISHED_GOODS

	CreatedAt time.Time `gorm:"index:idx_category_created,sort:desc"`
	UpdatedAt time.Time
}

type ProductImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;not null"`
	URL       string    `gorm:"type:varchar(255);not null"`
}

type Transaction struct {
	ID        uuid.UUID `gorm:"primary_key"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index"`
	Product   Product   `gorm:"foreignKey:ProductID"`
	BuyerID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Buyer     user.User `gorm:"foreignKey:BuyerID"`

	Quantity   int `gorm:"not null;default:1"` 
	TotalPrice int `gorm:"not null"`           

	Status string `gorm:"type:varchar(20);not null;index"`

	OfferID *uuid.UUID `gorm:"type:uuid;index"`
	SourceTransactionID *uuid.UUID `gorm:"type:uuid"`

	LogisticType *string `gorm:"type:varchar(20)"`
	MeetupLat    *float64
	MeetupLng    *float64 
	QRCode       *string  `gorm:"type:varchar(100);unique"` 

	ExpiresAt   *time.Time 
	ConfirmedAt *time.Time
	CompletedAt *time.Time
	CancelledAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
