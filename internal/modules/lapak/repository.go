package lapak

import (
	"context"
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
}


// == Product Repository ==
type CategoryRepository struct {
  db *gorm.DB
}
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
  return &CategoryRepository{db: db}
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
    err := cr.db.WithContext(ctx).Where("branch = ?", branch).Pluck("id", &categoryIDs).Error
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
  err := pr.db.WithContext(ctx).Preload("Image").Preload("user").Preload("Category").Where("id = ?", productID).First(&product).Error
  if err != nil {
    return nil, err
  }
  return &product, nil
}

func (pr *ProductRepository) FindList(ctx context.Context, categoryIDs []uuid.UUID, minPrice, maxPrice *int, cursor time.Time, limit int) ([]Product, error) {
  var product []Product

  db := pr.db.WithContext(ctx).Order("created_at desc").Limit(limit).Preload("User").Preload("Images").Preload("Category")

  if maxPrice == nil {
    db = db.Where("price < ?", maxPrice)
  }
  if minPrice == nil {
    db = db.Where("price < ?", minPrice)
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
  err := pr.db.WithContext(ctx).Model(&Product{}).Save(product).Error
  
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
