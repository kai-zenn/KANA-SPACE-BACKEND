package chat

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

func toMessageResponse(m *Message) *MessageResponse {
	return &MessageResponse{
		ID: m.ID, 
		SenderID: m.SenderID, 
		Type: m.Type, 
		Content: m.Content,
		OfferPrice: m.OfferPrice, 
		OfferStatus: m.OfferStatus, 
		CreatedAt: m.CreatedAt,
	}
}

var (
	ErrNotParticipant        = errors.New("bukan bagian dari percakapan ini")
	ErrOfferNotFound         = errors.New("offer tidak ditemukan")
	ErrOfferNotPending       = errors.New("offer sudah direspon sebelumnya")
	ErrCannotRespondOwnOffer = errors.New("tidak bisa merespon tawaran sendiri")
	ErrCannotOfferOnFreeItem = errors.New("tidak bisa menawar harga pada barang hibah/gratis")
	ErrOfferPriceRequired    = errors.New("offer_price wajib diisi untuk tipe pesan OFFER")
	ErrInvalidOfferPrice     = errors.New("harga tawaran tidak valid")
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
	return &MessageUseCase{
		cr: cr,
		mr: mr,
		la: la,
	}
}

func (mu *MessageUseCase) SendMessage(ctx context.Context, conversationID uuid.UUID, req CreateMessageRequest) (*MessageResponse, error) {
  conv, err := mu.cr.FindByID(ctx, conversationID)
  if err != nil {
    return nil, err
  }

  if conv.BuyerID != req.SenderID && conv.SellerID != req.SenderID {
		return nil, ErrNotParticipant
	}

	if req.Type == MessageTypeOffer {
		if req.OfferPrice == nil {
			return nil, ErrOfferPriceRequired
		}
		info, err := mu.la.GetProductInfo(ctx, conv.ProductID)
		if err != nil {
			return nil, err
		}
		if info.ListingType == "HIBAH" {
			return nil, ErrCannotOfferOnFreeItem
		}
		if err := mu.la.ValidateOfferPrice(ctx, conv.ProductID, *req.OfferPrice); err != nil {
			return nil, ErrInvalidOfferPrice
		}
	}

	msg := &Message{
		ID:          uuid.New(),
		ConversationID: conversationID,
		SenderID:    req.SenderID,
		Type:        req.Type,
		Content:     req.Content,
		OfferPrice:  req.OfferPrice,
	}

	if req.Type == MessageTypeOffer {
		msg.OfferPrice = req.OfferPrice
		status := OfferStatusPending
		msg.OfferStatus = &status
	}

	if err := mu.mr.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	return toMessageResponse(msg), nil
}

func (mu *MessageUseCase) GetMessages(ctx context.Context, conversationID, requesterID uuid.UUID, param MessageListQueryParam) (*MessageListResponse, error) {
  conv, err := mu.cr.FindByID(ctx, conversationID)
  if err != nil {
    return nil, err
  }

  if conv.BuyerID != requesterID && conv.SellerID != requesterID {
		return nil, ErrNotParticipant
	}

	limit := param.Limit
	if limit <= 0 || limit > 20 {
	  limit = 10
	}

	var cursor time.Time
	if param.Cursor != "" {
	  cleanCursor := strings.ReplaceAll(param.Cursor, " ", "+")
		var err error
		cursor, err = time.Parse(time.RFC3339, cleanCursor)
		if err != nil {
			return nil, err
		}
	}

	messages, err := mu.mr.FindByConversationID(ctx, conversationID, cursor, limit)

	responses := make([]MessageResponse, len(messages))
	for i, m := range messages {
		responses[i] = *toMessageResponse(&m)
	}

	var nextCursor *time.Time
	if len(messages) == limit {
		last := messages[len(messages)-1].CreatedAt
		nextCursor = &last
	}

	return &MessageListResponse{Messages: responses, NextCursor: nextCursor}, nil
}

func (mu *MessageUseCase) RespondToOffer(ctx context.Context, req RespondOfferRequest) (*MessageResponse, error) {
  msg, err := mu.mr.FindByID(ctx, req.MessageID)
  if err != nil {
    return nil, err
  }

  if msg.Type != MessageTypeOffer || msg.OfferStatus == nil {
    return nil, ErrOfferNotFound
  }

  if *msg.OfferStatus != OfferStatusPending {
    return nil, ErrOfferNotPending
  }

  if msg.SenderID == req.ResponderID {
    return nil, ErrNotParticipant
  }

  conv, err := mu.cr.FindByID(ctx, msg.ConversationID)
  if err != nil {
    return nil, err
  }

 	if conv.BuyerID != req.ResponderID && conv.SellerID != req.ResponderID {
		return nil, ErrNotParticipant
	}
 
	status := OfferStatusRejected
	if req.Accept {
		status = OfferStatusAccepted
	}
	msg.OfferStatus = &status
 
	if err := mu.mr.UpdateMessage(ctx, msg); err != nil {
		return nil, err
	}
	
	return toMessageResponse(msg), nil
}
