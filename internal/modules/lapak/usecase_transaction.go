package lapak

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	"github.com/google/uuid"

	"KANA-SPACE-BACKEND/internal/modules/user"
)

var (
	ErrProductNotAvailable  = errors.New("produk sudah tidak tersedia")
	ErrInsufficientStock    = errors.New("stok tidak cukup")
	ErrTransactionForbidden = errors.New("bukan pihak yang terlibat di transaksi ini")
	ErrInvalidTransactionStatus = errors.New("status transaksi tidak sesuai buat aksi ini")
	ErrUserFrozen = errors.New("akun sedang dibekukan dari klaim Rp0")
	ErrOfferNotAccepted = errors.New("offer belum di-accept penjual")
	ErrNotOfferOwner = errors.New("bukan pemilik offer ini")
	ErrProductLocked = errors.New("Produk sudah di Checkout (locked)")
	ErrOfferRejected = errors.New("offer sudah di-reject oleh penjual")
)

const selfPickupRadiusMeters = 4000

func haversineDistanceMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusMeters = 6371000
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}

func fuzzCoordinate(lat, lng float64) (float64, float64) {
	const precision = 1000.0
	return math.Round(lat*precision) / precision, math.Round(lng*precision) / precision
}

func toTransactionResponse(tx *Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID: tx.ID, ProductID: tx.ProductID, Quantity: tx.Quantity, TotalPrice: tx.TotalPrice,
		Status: tx.Status, LogisticType: tx.LogisticType, MeetupLat: tx.MeetupLat, MeetupLng: tx.MeetupLng,
		QRCode: tx.QRCode, ExpiresAt: tx.ExpiresAt, CreatedAt: tx.CreatedAt,
	}
}

func (tu *TransactionUseCase) lockProductForPickup(
	ctx context.Context, product *Product, buyerID uuid.UUID, price int,
	buyerLat, buyerLng float64, offerID *uuid.UUID,
) (*Transaction, error) {
	tx := &Transaction{
		ID: uuid.New(), ProductID: product.ID, BuyerID: buyerID,
		Quantity: 1, TotalPrice: price, OfferID: offerID,
	}

	distance := haversineDistanceMeters(buyerLat, buyerLng, product.Latitude, product.Longitude)
	if distance <= selfPickupRadiusMeters {
		logisticType := LogisticTypeSelfPickup
		tx.LogisticType = &logisticType
		fuzzedLat, fuzzedLng := fuzzCoordinate(product.Latitude, product.Longitude)
		tx.MeetupLat, tx.MeetupLng = &fuzzedLat, &fuzzedLng
	} else {
		logisticType := LogisticType3PL
		tx.LogisticType = &logisticType
	}

	qrCode := uuid.New().String()
	tx.QRCode = &qrCode
	expiresAt := time.Now().Add(24 * time.Hour)
	tx.ExpiresAt = &expiresAt
	tx.Status = TransactionStatusLocked

	product.Status = ProductStatusLocked
	if err := tu.pr.UpdateProduct(ctx, product); err != nil {
		return nil, err
	}
	if err := tu.tr.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}


type IChatAdapter interface {
	GetOfferInfo(ctx context.Context, offerID uuid.UUID) (*OfferInfo, error)
}

type OfferInfo struct {
	OfferID   uuid.UUID
	ProductID uuid.UUID
	SellerID  uuid.UUID
	BuyerID   uuid.UUID
	Price     int
	Status    string
}

type ITransactionUseCase interface {
	NewTransaction(ctx context.Context, req CreateTransactionRequest, buyerLat, buyerLng float64) (*TransactionResponse, error)
	ConfirmTransaction(ctx context.Context, transactionID, requesterID uuid.UUID) error
	CompleteViaQR(ctx context.Context, qrCode string, scannerID uuid.UUID) error
	CompleteManual(ctx context.Context, transactionID, requesterID uuid.UUID) error
	CancelTransaction(ctx context.Context, transactionID, requesterID uuid.UUID) error
	CheckoutFromOffer(ctx context.Context, offerID, buyerID uuid.UUID, buyerLat, buyerLng float64) (*TransactionResponse, error)
}

type TransactionUseCase struct {
	tr ITransactionRepository
	pr IProductRepository
	ur user.IUserRepository
	ca IChatAdapter
}

func NewTransactionUseCase(tr ITransactionRepository, pr IProductRepository, ur user.IUserRepository, ca IChatAdapter) ITransactionUseCase {
	return &TransactionUseCase{
		tr: tr,
		pr: pr,
		ur: ur,
		ca: ca,
	}
}

func (tu *TransactionUseCase) NewTransaction(ctx context.Context, req CreateTransactionRequest, buyerLat, buyerLng float64) (*TransactionResponse, error) {
  buyer, err := tu.ur.GetByID(ctx, req.BuyerID)
  if err != nil {
    return nil, err
  }
  if buyer.ClaimFreezeUntil != nil && buyer.ClaimFreezeUntil.After(time.Now()) {
    return nil, ErrUserFrozen
  }
  
  product, err := tu.pr.FindByID(ctx, req.ProductID)
  if err != nil {
    return nil, err
  }
  if product.Status != ProductStatusAvailable {
    return nil, ErrProductNotAvailable
  }

  tx := &Transaction{
    ID: uuid.New(),
    ProductID: req.ProductID,
    BuyerID: req.BuyerID,
  }

  switch product.Category.Branch{
    case CategoryBranchRawMaterial:
      tx.Quantity = 1
      tx.TotalPrice = product.Price
      
      distance := haversineDistanceMeters(buyerLat, buyerLng, product.Latitude, product.Longitude)
      if distance <= selfPickupRadiusMeters {
        logisticType := LogisticTypeSelfPickup
        tx.LogisticType = &logisticType
        fuzzedLat, fuzzedLng := fuzzCoordinate(product.Latitude, product.Longitude)
        tx.MeetupLat = &fuzzedLat
        tx.MeetupLng = &fuzzedLng
      } else {
        logisticeType := LogisticType3PL
        tx.LogisticType = &logisticeType
      }

      qrCode := uuid.New().String()
      tx.QRCode = &qrCode
      expiresAt := time.Now().Add(24 * time.Hour)
      tx.ExpiresAt = &expiresAt
      tx.Status = TransactionStatusLocked

      product.Status = ProductStatusLocked
      if err := tu.pr.UpdateProduct(ctx, product); err != nil {
        return nil, err
      }

     	tx, err := tu.lockProductForPickup(ctx, product, req.BuyerID, product.Price, buyerLat, buyerLng, nil)
     	if err != nil {
      		return nil, err
     	}
     	return toTransactionResponse(tx), nil
      
    case CategoryBranchFinishedGoods:
      if product.Stock == nil || *product.Stock < req.Quantity {
        return nil, ErrInsufficientStock
      }
      tx.Quantity = req.Quantity
      tx.TotalPrice = product.Price * req.Quantity

      logisticType := LogisticTypeArrangedOffline
      tx.LogisticType = &logisticType
      tx.Status = TransactionStatusPending

      newStock := *product.Stock - req.Quantity
      product.Stock = &newStock
      if newStock == 0 {
        product.Status = ProductStatusInactive
      }
      if err := tu.pr.UpdateProduct(ctx, product); err != nil {
        return nil, err
      }
  }

  if err := tu.tr.CreateTransaction(ctx, tx); err != nil {
    return nil, err
  }

  return toTransactionResponse(tx), nil
}

func (tu *TransactionUseCase) ConfirmTransaction(ctx context.Context, transactionID, requesterID uuid.UUID) error {
 	tx, err := tu.tr.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx.Product.UserID != requesterID {
		return ErrTransactionForbidden
	}
	if tx.Status != TransactionStatusPending {
		return ErrInvalidTransactionStatus
	}
 
	now := time.Now()
	tx.Status = TransactionStatusConfirmed
	tx.ConfirmedAt = &now
	return tu.tr.UpdateTransaction(ctx, tx)
}

func (tu *TransactionUseCase) CompleteViaQR(ctx context.Context, qrCode string, scannerID uuid.UUID) error {
	tx, err := tu.tr.FindByQRCode(ctx, qrCode)
	if err != nil {
		return err
	}
	if tx.Product.UserID != scannerID {
		return ErrTransactionForbidden
	}
	if tx.Status != TransactionStatusLocked {
		return ErrInvalidTransactionStatus
	}

	now := time.Now()
	tx.Status = TransactionStatusCompleted
	tx.CompletedAt = &now
	if err := tu.tr.UpdateTransaction(ctx, tx); err != nil {
		return err
	}

	tx.Product.Status = ProductStatusCompleted
	return tu.pr.UpdateProduct(ctx, &tx.Product)
}

func (tu *TransactionUseCase) CompleteManual(ctx context.Context, transactionID, requesterID uuid.UUID) error {
	tx, err := tu.tr.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx.Product.UserID != requesterID && tx.BuyerID != requesterID {
		return ErrTransactionForbidden
	}
	if tx.Status != TransactionStatusConfirmed {
		return ErrInvalidTransactionStatus
	}

	now := time.Now()
	tx.Status = TransactionStatusCompleted
	tx.CompletedAt = &now
	return tu.tr.UpdateTransaction(ctx, tx)
}

func (tu *TransactionUseCase) CancelTransaction(ctx context.Context, transactionID, requesterID uuid.UUID) error {
	tx, err := tu.tr.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx.Product.UserID != requesterID && tx.BuyerID != requesterID {
		return ErrTransactionForbidden
	}
	if tx.Status != TransactionStatusPending && tx.Status != TransactionStatusConfirmed {
		return ErrInvalidTransactionStatus
	}

	now := time.Now()
	tx.Status = TransactionStatusCancelled
	tx.CancelledAt = &now
	if err := tu.tr.UpdateTransaction(ctx, tx); err != nil {
		return err
	}

	restoredStock := *tx.Product.Stock + tx.Quantity
	tx.Product.Stock = &restoredStock
	tx.Product.Status = ProductStatusAvailable
	return tu.pr.UpdateProduct(ctx, &tx.Product)
}

func (tu *TransactionUseCase) ExpireStaleTransactions(ctx context.Context) error {
	expired, err := tu.tr.FindExpiredLocked(ctx, time.Now())
	if err != nil {
		return err
	}

	for _, tx := range expired {
		tx.Status = TransactionStatusExpired
		if err := tu.tr.UpdateTransaction(ctx, &tx); err != nil {
					log.Printf("Gagal update status transaksi %s: %v\n", tx.ID, err)
				}

		tx.Product.Status = ProductStatusAvailable
		if err := tu.pr.UpdateProduct(ctx, &tx.Product); err != nil {
					log.Printf("Gagal mengembalikan status produk %s: %v\n", tx.Product.ID, err)
				}

		buyer, err := tu.ur.GetByID(ctx, tx.BuyerID)
		if err != nil {
			continue
		}
		buyer.StrikeCount++
		if buyer.StrikeCount >= 3 {
			freezeUntil := time.Now().Add(30 * 24 * time.Hour)
			buyer.ClaimFreezeUntil = &freezeUntil
		}
		errUpdate := tu.ur.UpdateUser(ctx, buyer.ID, map[string]interface{}{
					"strike_count":       buyer.StrikeCount,
					"claim_freeze_until": buyer.ClaimFreezeUntil,
				})
		
		if errUpdate != nil {
			log.Printf("Gagal update strike user %s: %v\n", buyer.ID, errUpdate)
		}
	}
	return nil
}

func (tu *TransactionUseCase) CheckoutFromOffer(ctx context.Context, offerID, buyerID uuid.UUID, buyerLat, buyerLng float64) (*TransactionResponse, error) {
	offer, err := tu.ca.GetOfferInfo(ctx, offerID)
	if err != nil {
		return nil, err
	}
	if offer.Status != "ACCEPTED" {
		return nil, ErrOfferNotAccepted
	}
	if offer.BuyerID != buyerID {
		return nil, ErrNotOfferOwner
	}

	product, err := tu.pr.FindByID(ctx, offer.ProductID)
	if err != nil {
		return nil, err
	}
	if product.Status == ProductStatusLocked {
		return nil, ErrProductLocked
	}
	if product.Status != ProductStatusAvailable {
		return nil, ErrProductNotAvailable
	}

	tx, err := tu.lockProductForPickup(ctx, product, buyerID, offer.Price, buyerLat, buyerLng, &offerID)
	if err != nil {
		return nil, err
	}

	return toTransactionResponse(tx), nil
}
