package chat

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotParticipant        = errors.New("bukan bagian dari percakapan ini")
	ErrOfferNotFound         = errors.New("offer tidak ditemukan")
	ErrOfferNotPending       = errors.New("offer sudah direspon sebelumnya")
	ErrCannotRespondOwnOffer = errors.New("tidak bisa merespon tawaran sendiri")
	ErrCannotOfferOnFreeItem = errors.New("tidak bisa menawar harga pada barang hibah/gratis")
	ErrOfferPriceRequired    = errors.New("offer_price wajib diisi untuk tipe pesan OFFER")
)

type ILapakAdapter interface {
	GetProductInfo(ctx context.Context, productID uuid.UUID) (*ProductInfo, error)
	ValidateOfferPrice(ctx context.Context, productID uuid.UUID, price int) error
}

type ProductInfo struct {
	ProductID   uuid.UUID
	SellerID    uuid.UUID
	Status      string
	ListingType string 
}

type IMessageUseCase interface {
	SendMessage(ctx context.Context, conversationID uuid.UUID, req CreateMessageRequest) (*MessageResponse, error)
	GetMessages(ctx context.Context, conversationID, requesterID uuid.UUID, param MessageListQueryParam) (*MessageListResponse, error)
	RespondToOffer(ctx context.Context, req RespondOfferRequest) (*MessageResponse, error)
}

type MessageUseCase struct {
	cr IConversationRepository
	mr IMessageRepository
	la ILapakAdapter
}

func NewMessageUseCase(cr IConversationRepository, mr IMessageRepository, la ILapakAdapter) IMessageUseCase {
	return &MessageUseCase{cr: cr, mr: mr, la: la}
}
