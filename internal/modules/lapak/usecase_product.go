package lapak

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"KANA-SPACE-BACKEND/internal/modules/user"
	"KANA-SPACE-BACKEND/internal/pkgs/storage"
)

var (
	ErrOnlySellerCanListProduct = errors.New("cuma seller yang boleh bikin listing produk")
	ErrInvalidPriceRange        = errors.New("harga di luar rentang yang diizinkan")
	ErrProductForbidden         = errors.New("bukan pemilik produk ini")
	ErrCategoryMustBeSubcategory = errors.New("harus pilih subkategori, bukan kategori besar")
	ErrListingTypeMismatch      = errors.New("listing type nggak cocok buat kategori ini")
  ErrInvalidCategory         = errors.New("category_id tidak valid")
  ErrSelfDeclarationRequired = errors.New("self declaration tag wajib buat bahan baku")
  ErrPriceMustBePositive     = errors.New("harga harus lebih dari 0")
)

type NLPClientInterface interface {
	Embed(ctx context.Context, text string) (embedding []float64, model string, err error)
}

func ToProductSeller(u user.User) ProductSeller {
	return ProductSeller{ID: u.ID, Username: u.Username, ProfilePhotoLink: u.ProfilePhotoLink}
}

func ToCategoryResponse(c Category) CategoryResponse {
	return CategoryResponse{ID: c.ID, Name: c.Name, Slug: c.Slug, Branch: c.Branch}
}

type IProductUseCase interface {
	NewProduct(ctx context.Context, req CreateProductRequest, requesterRole string) (*ProductResponse, error)
	FindProductByID(ctx context.Context, productID uuid.UUID) (*ProductResponse, error)
	GetProductList(ctx context.Context, param ProductListQueryParam) (*ProductListResponse, error)
	UpdateProduct(ctx context.Context, req UpdateProductRequest) (*ProductResponse, error)
	DeleteProduct(ctx context.Context, productID, requesterID uuid.UUID, requesterRole string) error
}

type ProductUseCase struct {
	pr      IProductRepository
	cr      ICategoryRepository
	ur      user.IUserRepository
	nlp     NLPClientInterface
	storage storage.Interface
}

func NewProductUseCase(pr IProductRepository, cr ICategoryRepository, ur user.IUserRepository, nlp NLPClientInterface, storage storage.Interface) IProductUseCase {
	return &ProductUseCase{pr: pr, cr: cr, ur: ur, nlp: nlp, storage: storage}
}


func (pu *ProductUseCase) NewProduct(ctx context.Context, req CreateProductRequest, requesterRole string) (*ProductResponse, error) {
  if requesterRole != user.RoleSeller && requesterRole != user.RoleAdmin {
    return nil, ErrOnlySellerCanListProduct
  }

  categoryID, err := uuid.Parse(req.CategoryID)
  if err != nil {
    return nil, errors.New("category_id tidak valid")
  }

  category, err := pu.cr.FindByID(ctx, categoryID)
  if err != nil {
    return nil, err
  }

  if category.ParentID == nil {
    return nil, ErrCategoryMustBeSubcategory
  }

  var selfDeclarationTag *string
 	switch category.Branch {
  	case CategoryBranchRawMaterial:
  		if req.ListingType != ListingTypeHibah && req.ListingType != ListingTypeJualBorongan {
  			return nil, ErrListingTypeMismatch
  		}
  		if req.ListingType == ListingTypeJualBorongan {
  			if req.Price < 5000 || req.Price > 15000 {
  				return nil, ErrInvalidPriceRange
  			}
  		} else {
  			req.Price = 0
  		}
  		if req.SelfDeclarationTag == "" {
  			return nil, errors.New("self declaration tag wajib buat bahan baku")
  		}
  		tag := req.SelfDeclarationTag
  		selfDeclarationTag = &tag
  
  	case CategoryBranchFinishedGoods:
  		if req.ListingType != ListingTypeHibah && req.ListingType != ListingTypeDijual {
  			return nil, ErrListingTypeMismatch
  		}
  		if req.ListingType == ListingTypeDijual {
  			if req.Price <= 0 {
  				return nil, errors.New("harga harus lebih dari 0")
  			}
  		} else {
  			req.Price = 0
  		}
  		selfDeclarationTag = nil
  
  	default:
  		return nil, fmt.Errorf("branch kategori tidak dikenal: %s", category.Branch)
	}

	newPhotoUrl, err := pu.storage.UploadProductImages(ctx, req.Images)
	if err != nil {
	  return nil, err
	}

	defer func() {
	  if err != nil && len(newPhotoUrl) > 0 {
			  go func() {
					  _ = pu.storage.DeleteProductImages(context.Background(), newPhotoUrl)
					}()
			}
	}()

	productID := uuid.New()
	var images []ProductImage

	for _, url := range newPhotoUrl {
	  images = append(images, ProductImage{ID: uuid.New(), URL: url, ProductID: productID })
	}

	seller, err := pu.ur.GetByID(ctx, req.UserID)
	if err != nil {
	  return nil, err
	}

	product := &Product{
  	ID: productID, UserID: req.UserID, Title: req.Title, Description: req.Description,
  	CategoryID: category.ID, ListingType: req.ListingType, Price: req.Price, Stock: req.Stock,
  	SelfDeclarationTag: selfDeclarationTag, Status: ProductStatusAvailable,
  	Latitude: req.Latitude, Longitude: req.Longitude, Images: images,
	}

	err = pu.pr.CreateProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	return &ProductResponse{
		ID: product.ID, Seller: ToProductSeller(*seller), Title: product.Title,
		Description: product.Description, Category: ToCategoryResponse(*category),
		ListingType: product.ListingType, Price: product.Price, Stock: product.Stock,
		SelfDeclarationTag: product.SelfDeclarationTag, Status: product.Status,
		PhotoURLs: newPhotoUrl, CreatedAt: product.CreatedAt,
	}, nil
}

func (pu *ProductUseCase) FindProductByID(ctx context.Context, productID uuid.UUID) (*ProductResponse, error) {
  product, err := pu.pr.FindByID(ctx, productID)
  if err != nil {
    return nil, err
  }

  photoURLs := make([]string, len(product.Images))
  for i, img := range product.Images {
    photoURLs[i] = img.URL
  }

  return &ProductResponse{
    ID: productID,
    Seller: ToProductSeller(product.User),
    Title: product.Title,
    Description: product.Description,
    Category: ToCategoryResponse(product.Category),
    ListingType: product.ListingType,
    Price: product.Price,
    SelfDeclarationTag: product.SelfDeclarationTag,
    Status: product.Status,
    PhotoURLs: photoURLs,
    CreatedAt: product.CreatedAt,
  }, nil
}

func (pu *ProductUseCase) GetProductList(ctx context.Context, param ProductListQueryParam) (*ProductListResponse, error) {
 	var categoryIDs []uuid.UUID
 
	if param.CategoryID != "" {
		id, err := uuid.Parse(param.CategoryID)
		if err != nil {
			return nil, errors.New("category_id tidak valid")
		}
		categoryIDs = []uuid.UUID{id}
	} else if param.Branch != "" {
		ids, err := pu.cr.FindIDsByBranch(ctx, param.Branch)
		if err != nil {
			return nil, err
		}
		categoryIDs = ids
	}
	
  limit := param.Limit
  if limit <= 0 || limit > 20 {
    limit = 15
  }

  var cursor time.Time
  if param.Cursor != "" {
     var err error
     cursor, err = time.Parse(time.RFC3339, param.Cursor)
     if err != nil {
       return nil, errors.New("format cursor tidak valid, gunakan format RFC3339")
     }
  }

  products, err := pu.pr.FindList(ctx, categoryIDs, param.MinPrice, param.MaxPrice, cursor, limit)
  if err != nil {
    return nil, err
  }

  response := make([]ProductResponse, len(products))
  for i, p := range products {
    photoURLs := make([]string, len(p.Images))
    for j, img := range p.Images {
      photoURLs[j] = img.URL
    }
    response[i] = ProductResponse{
      ID: p.ID,
      Seller: ToProductSeller(p.User),
      Title: p.Title,
      Description: p.Description,
      Category: ToCategoryResponse(p.Category),
      ListingType: p.ListingType,
      Price: p.Price,
      SelfDeclarationTag: p.SelfDeclarationTag,
      Status: p.Status,
      PhotoURLs: photoURLs,
      CreatedAt: p.CreatedAt,
    }
  }

  var nextCursor *time.Time
  if len(products) == limit {
    last := products[len(products)-1].CreatedAt
    nextCursor = &last
  }

  return &ProductListResponse{
    Products: response,
    NextCursor: nextCursor,
  }, nil
}

func (pu *ProductUseCase) UpdateProduct(ctx context.Context, req UpdateProductRequest) (*ProductResponse, error) {
 	product, err := pu.pr.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if product.UserID != req.UserID {
		return nil, ErrProductForbidden
	}
 
	if req.Title != nil {
		product.Title = *req.Title
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		if product.ListingType == ListingTypeJualBorongan && (*req.Price < 5000 || *req.Price > 15000) {
			return nil, ErrInvalidPriceRange
		}
		if product.ListingType == ListingTypeDijual && *req.Price <= 0 {
			return nil, errors.New("harga harus lebih dari 0")
		}
		product.Price = *req.Price
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
 
	if err := pu.pr.UpdateProduct(ctx, product); err != nil {
		return nil, err
	}
 
	return pu.FindProductByID(ctx, product.ID)
}

func (pu *ProductUseCase) DeleteProduct(ctx context.Context, productID, requesterID uuid.UUID, requesterRole string) error {
	product, err := pu.pr.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.UserID != requesterID && requesterRole != user.RoleAdmin {
		return ErrProductForbidden
	}

	photoURLs := make([]string, len(product.Images))
	for i, img := range product.Images {
		photoURLs[i] = img.URL
	}

	if err := pu.pr.DeleteProduct(ctx, productID); err != nil {
		return err
	}

	go func() {
		_ = pu.storage.DeletePostImages(context.Background(), photoURLs)
	}()

	return nil
}
