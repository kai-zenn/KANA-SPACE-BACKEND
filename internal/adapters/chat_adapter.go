package adapters

import (
	"context"
	"errors"

	"KANA-SPACE-BACKEND/internal/modules/chat"
	"KANA-SPACE-BACKEND/internal/modules/lapak"

	"github.com/google/uuid"
)

type ChatAdapter struct {
	messageRepo chat.IMessageRepository
	conversationRepo chat.IConversationRepository
}

func NewChatAdapter(messageRepo chat.IMessageRepository, conversationRepo chat.IConversationRepository) lapak.IChatAdapter {
	return &ChatAdapter{
		messageRepo: messageRepo,
		conversationRepo: conversationRepo,
	}
}

func (ca *ChatAdapter) GetOfferInfo(ctx context.Context, offerID uuid.UUID) (*lapak.OfferInfo, error) {
	msg, err := ca.messageRepo.FindByID(ctx, offerID)
	if err != nil {
		return nil, errors.New("tawaran tidak ditemukan")
	}

	if msg.Type != chat.MessageTypeOffer || msg.OfferPrice == nil || msg.OfferStatus == nil {
		return nil, errors.New("pesan ini bukan sebuah tawaran harga")
	}

	conv, err := ca.conversationRepo.FindByID(ctx, msg.ConversationID)
	if err != nil {
		return nil, errors.New("data percakapan tidak ditemukan")
	}

	return &lapak.OfferInfo{
		OfferID:   msg.ID,
		ProductID: conv.ProductID,
		SellerID:  conv.SellerID,
		BuyerID:   conv.BuyerID,
		Price:     *msg.OfferPrice,
		Status:    *msg.OfferStatus,
	}, nil
}
