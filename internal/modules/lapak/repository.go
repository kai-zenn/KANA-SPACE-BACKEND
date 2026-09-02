package lapak

import (
	"KANA-SPACE-BACKEND/internal/pkgs/geo"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ICategoryRepository interface {
	FindAll(ctx context.Context) ([]Category, error)
	FindByID(ctx context.Context, categoryID uuid.UUID) (*Category, error)
	FindIDsByBranch(ctx context.Context, branch string) ([]uuid.UUID, error)
}

type IProductRepository interface {
	CreateProduct(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, productID uuid.UUID) (*Product, error)
	FindList(ctx context.Context, categoryIDs []uuid.UUID, minPrice, maxPrice *int, cursor time.Time, limit int) ([]Product, error)
	UpdateProduct(ctx context.Context, product *Product) error
	// UpdateEmbedding(ctx context.Context, productID uuid.UUID, embedding []float64, model string) error
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
	FindAvailableRawMaterialCandidates(ctx context.Context, lat, lng, radiusMeters float64) ([]ProductCandidate, error)
	FindNearby(ctx context.Context, params NearbyParams) ([]Product, error)
}

type ITransactionRepository interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	FindByID(ctx context.Context, transactionId uuid.UUID) (*Transaction, error)
	FindByQRCode(ctx context.Context, qrCode string) (*Transaction, error)
	FindExpiredLocked(ctx context.Context, now time.Time) ([]Transaction, error) // buat cron
	UpdateTransaction(ctx context.Context, tx *Transaction) error
	BulkExpireLocked(ctx context.Context) ([]uuid.UUID, error)
  CompleteIfLocked(ctx context.Context, id uuid.UUID) error
}


// == Category Repository ==
type CategoryRepository struct {
  db *gorm.DB
}
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
  return &CategoryRepository{
    db: db,
  }
}

func (cr *CategoryRepository) FindAll(ctx context.Context) ([]Category, error) {
  var category []Category

  err := cr.db.WithContext(ctx).Find(&category).Error
  if err != nil {
    return nil, err
  }

  return category, err
}

func (cr *CategoryRepository) FindByID(ctx context.Context, categoryID uuid.UUID) (*Category, error) {
    var category Category

    err := cr.db.WithContext(ctx).Where("id = ?", categoryID).First(&category).Error
    if err != nil {
      return nil, err
    }
    return &category, nil
}

func (cr *CategoryRepository) FindIDsByBranch(ctx context.Context, branch string) ([]uuid.UUID, error) {
    var categoryIDs []uuid.UUID
    err := cr.db.WithContext(ctx).Model(Category{}).Where("branch = ?", branch).Pluck("id", &categoryIDs).Error
    if err != nil {
      return nil, err
    }
    return categoryIDs, nil
}


// == Product Repository ==
type ProductRepository struct {
  db *gorm.DB
}
func NewProductRepository(db *gorm.DB) *ProductRepository {
  return &ProductRepository{db: db}
}

func (pr *ProductRepository) CreateProduct(ctx context.Context, product *Product) error {
  return pr.db.WithContext(ctx).Create(product).Error
}

func (pr *ProductRepository) FindByID(ctx context.Context, productID uuid.UUID) (*Product, error) {
  var product Product
  err := pr.db.WithContext(ctx).Preload("Images").Preload("User").Preload("Category").Where("id = ?", productID).First(&product).Error
  if err != nil {
    return nil, err
  }
  return &product, nil
}

func (pr *ProductRepository) FindList(ctx context.Context, categoryIDs []uuid.UUID, minPrice, maxPrice *int, cursor time.Time, limit int) ([]Product, error) {
  var product []Product

  db := pr.db.WithContext(ctx).Order("created_at desc").Limit(limit).Preload("User").Preload("Images").Preload("Category")

  if len(categoryIDs) > 0 {
		db = db.Where("category_id IN ?", categoryIDs)
	}
	
  if maxPrice != nil {
    db = db.Where("price <= ?", maxPrice)
  }
  if minPrice != nil {
    db = db.Where("price >= ?", minPrice)
  }
  
  if !cursor.IsZero() {
    db = db.Where("created_at < ?", cursor)
  }

  err := db.Find(&product).Error
  if err != nil {
    return nil, err
  }

  return product, nil
}

func (pr *ProductRepository) UpdateProduct(ctx context.Context, product *Product) error {
  err := pr.db.WithContext(ctx).Model(&Product{}).Where("id = ?", product.ID).Save(product).Error
  
  if err != nil {
    return err
  }
  
  return nil
}

// func (pr *ProductRepository) UpdateEmbedding(ctx context.Context, productID uuid.UUID, embedding []float64, model string) error {
  
// }

func (pr *ProductRepository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
  err := pr.db.
    WithContext(ctx).
    Where("id = ?", productID).
    Delete(&Product{}).
    Error

  if err != nil {
    return err
  }

  return nil
}

func (pr *ProductRepository) FindAvailableRawMaterialCandidates(
	ctx context.Context, lat, lng, radiusMeters float64,
) ([]ProductCandidate, error) {
	
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusMeters)
	var candidates []ProductCandidate
	
	err := pr.db.WithContext(ctx).Raw(`
		SELECT
			products.id,
			products.title,
			products.description,
			products.embedding,
			(
				6371000 * 2 * asin(
					sqrt(
						LEAST(1,
							power(sin(radians(latitude - ?) / 2), 2) +
							cos(radians(?)) * cos(radians(latitude)) *
							power(sin(radians(longitude - ?) / 2), 2)
						)
					)
				)
			) AS distance_meters
		FROM products
		INNER JOIN categories ON categories.id = products.category_id
		WHERE products.status = 'AVAILABLE'
		  AND categories.branch = 'RAW_MATERIAL' 
		  AND products.embedding IS NOT NULL
		  AND products.latitude BETWEEN ? AND ? 
		  AND products.longitude BETWEEN ? AND ?
		HAVING distance_meters <= ?
		ORDER BY distance_meters ASC
		LIMIT 50
	`, lat, lat, lng, minLat, maxLat, minLng, maxLng, radiusMeters).Scan(&candidates).Error

	return candidates, err
}

func (pr *ProductRepository) FindNearby(ctx context.Context, params NearbyParams) ([]Product, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(params.Lat, params.Lng, params.RadiusMeters)

	const query = `
		SELECT * FROM (
			SELECT products.*,
				6371000 * 2 * asin(
					sqrt(
						LEAST(1,
							power(sin(radians(latitude - ?) / 2), 2) +
							cos(radians(?)) * cos(radians(latitude)) *
							power(sin(radians(longitude - ?) / 2), 2)
						)
					)
				) AS distance_meters
			FROM products
			WHERE status = 'AVAILABLE'
			  AND latitude BETWEEN ? AND ? 
			  AND longitude BETWEEN ? AND ?
		) nearby
		WHERE distance_meters <= ? 
		ORDER BY distance_meters ASC
		LIMIT ? OFFSET ?;
	`

	var products []Product
	err := pr.db.WithContext(ctx).Raw(query,
		params.Lat, params.Lat, params.Lng,
		minLat, maxLat, minLng, maxLng,
		params.RadiusMeters,
		params.Limit, params.Offset,
	).Scan(&products).Error

	return products, err
}

// == Transaction Repository ==
type TransactionRepository struct {
  db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository{
  return &TransactionRepository{
    db: db,
  }
}

func (tr *TransactionRepository) CreateTransaction(ctx context.Context, tx *Transaction) error {
  err := tr.db.WithContext(ctx).Create(tx).Error
  if err != nil {
    return err
  }
  return nil
}

func (tr *TransactionRepository) FindByID(ctx context.Context, transactionId uuid.UUID) (*Transaction, error) {
  var tx Transaction
  err := tr.db.WithContext(ctx).Preload("Product").Where("id = ?", transactionId).First(&tx).Error
  if err != nil {
    return nil, err
  }
  return &tx, nil
}

func (tr *TransactionRepository) FindByQRCode(ctx context.Context, qrCode string) (*Transaction, error) {
  var tx Transaction
  err := tr.db.WithContext(ctx).Preload("Product").Where("qr_code = ?", qrCode).First(&tx).Error
  if err != nil {
    return nil, err
  }
  return &tx, nil
}

func (tr *TransactionRepository) FindExpiredLocked(ctx context.Context, now time.Time) ([]Transaction, error) {
  var txs []Transaction
  err := tr.db.WithContext(ctx).Preload("Product").Where("expires_at < ? AND status = ?", now, TransactionStatusLocked).Find(&txs).Error
  if err != nil {
    return nil, err
  }
  return txs, nil
}

func (tr *TransactionRepository) UpdateTransaction(ctx context.Context, tx *Transaction) error{
  err := tr.db.WithContext(ctx).Model(&Transaction{}).Where("id = ?", tx.ID).Save(tx).Error
  
  if err != nil {
    return err
  }
  
  return nil
}

func (tr *TransactionRepository) CompleteIfLocked(ctx context.Context, id uuid.UUID) error {
    now := time.Now()
    result := tr.db.WithContext(ctx).Model(&Transaction{}).
        Where("id = ? AND status = ?", id, TransactionStatusLocked).
        Updates(map[string]interface{}{
            "status":       TransactionStatusCompleted, 
            "completed_at": now,
        })

    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("transaksi sudah tidak valid (kadaluarsa atau sudah diproses)")
    }
    return nil
}

func (tr *TransactionRepository) BulkExpireLocked(ctx context.Context) ([]uuid.UUID, error) {
    const query = `
        WITH expired AS (
            UPDATE transactions
            SET status = 'EXPIRED', updated_at = NOW()
            WHERE status = 'LOCKED'
              AND expires_at <= NOW()
            RETURNING id, product_id
        ),
        released AS (
            UPDATE products
            SET status = 'AVAILABLE', updated_at = NOW()
            FROM expired
            WHERE products.id = expired.product_id
        )
        SELECT id FROM expired;
    `
    var ids []uuid.UUID
    err := tr.db.WithContext(ctx).Raw(query).Scan(&ids).Error
    return ids, err
}
