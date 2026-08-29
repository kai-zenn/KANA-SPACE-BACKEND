package lapak

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type CategoryResponse struct {
	ID            uuid.UUID          `json:"id"`
	Name          string             `json:"name"`
	Slug          string             `json:"slug"`
	Branch        string             `json:"branch"`
	Subcategories []CategoryResponse `json:"subcategories,omitempty"`
}

type CreateProductRequest struct {
	UserID             uuid.UUID               `json:"-"`
	Title              string                  `form:"title" binding:"required"`
	Description        string                  `form:"description" binding:"required"`
	CategoryID         string                  `form:"category_id" binding:"required,uuid"`
	ListingType        string                  `form:"listing_type" binding:"required,oneof=HIBAH JUAL_BORONGAN DIJUAL"`
	Price              int                     `form:"price"`
	Stock              *int                    `form:"stock"`
	SelfDeclarationTag string                  `form:"self_declaration_tag"`
	Latitude           float64                 `form:"latitude" binding:"required"`
	Longitude          float64                 `form:"longitude" binding:"required"`
	Images             []*multipart.FileHeader `form:"images" binding:"required,min=1,max=4"`
}

type ProductSeller struct {
	ID               uuid.UUID `json:"id"`
	Username         string    `json:"username"`
	ProfilePhotoLink string    `json:"profile_photo_link"`
}

type ProductResponse struct {
	ID                 uuid.UUID        `json:"id"`
	Seller             ProductSeller    `json:"seller"`
	Title              string           `json:"title"`
	Description        string           `json:"description"`
	Category           CategoryResponse `json:"category"`
	ListingType        string           `json:"listing_type"`
	Price              int              `json:"price"`
	Stock              *int             `json:"stock,omitempty"`
	SelfDeclarationTag *string          `json:"self_declaration_tag,omitempty"`
	Status             string           `json:"status"`
	PhotoURLs          []string         `json:"photo_urls"`
	CreatedAt          time.Time        `json:"created_at"`
}

type ProductListQueryParam struct {
	Branch     string `form:"branch"`
	CategoryID string `form:"category_id"`
	MinPrice   *int   `form:"min_price"`
	MaxPrice   *int   `form:"max_price"`
	Cursor     string `form:"cursor"`
	Limit      int    `form:"limit"`
}

type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	NextCursor *time.Time        `json:"next_cursor,omitempty"`
}

type UpdateProductRequest struct {
	ProductID   uuid.UUID `json:"-"`
	UserID      uuid.UUID `json:"-"`
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Price       *int      `json:"price"`
	Status      *string   `json:"status" binding:"omitempty,oneof=AVAILABLE INACTIVE"`
}

type CreateTransactionRequest struct {
	ProductID uuid.UUID `json:"-"`
	BuyerID   uuid.UUID `json:"-"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

type TransactionResponse struct {
	ID           uuid.UUID  `json:"id"`
	ProductID    uuid.UUID  `json:"product_id"`
	Quantity     int        `json:"quantity"`
	TotalPrice   int        `json:"total_price"`
	Status       string     `json:"status"`
	LogisticType *string    `json:"logistic_type,omitempty"`
	MeetupLat    *float64   `json:"meetup_lat,omitempty"`
	MeetupLng    *float64   `json:"meetup_lng,omitempty"`
	QRCode       *string    `json:"qr_code,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CheckoutFromOfferRequest struct {
	OfferID  uuid.UUID `json:"offer_id" binding:"required"`
	BuyerLat float64   `json:"buyer_lat" binding:"required"`
	BuyerLng float64   `json:"buyer_lng" binding:"required"`
}
