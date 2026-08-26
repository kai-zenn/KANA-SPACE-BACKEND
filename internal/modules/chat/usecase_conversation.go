package chat

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrCannotChatOwnProduct = errors.New("tidak bisa membuka chat di produk sendiri")
)

type IConversationUseCase interface {
	GetOrCreateConversation(ctx context.Context, productID, requesterID uuid.UUID) (*ConversationResponse, error)
	ListConversations(ctx context.Context, requesterID uuid.UUID) (*ConversationListResponse, error)
}

type ConversationUseCase struct {
	cr IConversationRepository
	la ILapakAdapter
}

func NewConversationUseCase(cr IConversationRepository, la ILapakAdapter) IConversationUseCase {
	return &ConversationUseCase{cr: cr, la: la}
}
