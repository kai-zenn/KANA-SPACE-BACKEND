package adapters

import (
	"context"
	"errors"

	"KANA-SPACE-BACKEND/internal/modules/lapak"
	"KANA-SPACE-BACKEND/internal/modules/chat"

	"github.com/google/uuid"
)

type LapakAdapter struct {
	productRepo lapak.IProductRepository
}

func NewLapakAdapter(productRepo lapak.IProductRepository) chat.ILapakAdapter {
	return &LapakAdapter{
		productRepo: productRepo,
	}
}

func (la *LapakAdapter) GetProductInfo(ctx context.Context, productID uuid.UUID) (*chat.ProductInfo, error) {
	product, err := la.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, errors.New("produk tidak ditemukan")
	}

	return &chat.ProductInfo{
		ProductID:   product.ID,
		SellerID:    product.UserID,
		Status:      product.Status,
		ListingType: product.ListingType,
	}, nil
}

func (la *LapakAdapter) ValidateOfferPrice(ctx context.Context, productID uuid.UUID, price int) error {
	product, err := la.productRepo.FindByID(ctx, productID)
	if err != nil {
		return errors.New("produk tidak ditemukan")
	}

	if price <= 0 {
		return errors.New("harga tawaran harus lebih besar dari 0")
	}
	if price > product.Price {
		return errors.New("harga tawaran tidak boleh melebihi harga asli produk")
	}

	return nil
}
