package chat

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrCannotChatOwnProduct = errors.New("tidak bisa membuka chat di produk sendiri")
)

func toConversationResponse(c *Conversation) *ConversationResponse {
	return &ConversationResponse{
		ID: c.ID, 
		ProductID: c.ProductID, 
		SellerID: c.SellerID, 
		BuyerID: c.BuyerID, 
		CreatedAt: c.CreatedAt,
	}
}

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

func (cu *ConversationUseCase) GetOrCreateConversation(ctx context.Context, productID, requesterID uuid.UUID) (*ConversationResponse, error) {
	existingConv, err := cu.cr.FindByProductAndBuyer(ctx, productID, requesterID)
	if err == nil && existingConv != nil {
		return toConversationResponse(existingConv), nil
	}

	info, err := cu.la.GetProductInfo(ctx, productID)
	if err != nil {
		return nil, err
	}
	if info.SellerID == requesterID {
		return nil, ErrCannotChatOwnProduct
	}

	conv := &Conversation{
		ID: uuid.New(),
		ProductID: info.ProductID,
		SellerID: info.SellerID,
		BuyerID: requesterID,
	}
	if err := cu.cr.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}
	return toConversationResponse(conv), nil
}

func (cu *ConversationUseCase) ListConversations(ctx context.Context, requesterID uuid.UUID) (*ConversationListResponse, error) {
	conversations, err := cu.cr.FindByParticipant(ctx, requesterID)
	if err != nil {
		return nil, err
	}
	responses := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		responses[i] = *toConversationResponse(&c)
	}
	return &ConversationListResponse{Conversations: responses}, nil
}
