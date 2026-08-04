package lapak

import (
	"context"

	"github.com/google/uuid"
)

type ICategoryUseCase interface {
	GetCategoryTree(ctx context.Context) ([]CategoryResponse, error)
}

type CategoryUseCase struct {
	cr ICategoryRepository
}

func NewCategoryUseCase(cr ICategoryRepository) ICategoryUseCase {
	return &CategoryUseCase{cr: cr}
}

func sameParent(a, b *uuid.UUID) bool {
  if a == nil && b == nil {
    return true
  }
  if a == nil || b == nil {
    return false
  }
  return *a == *b
}
 
func buildCategoryTree(all []Category, parentID *uuid.UUID) []CategoryResponse {
  var result []CategoryResponse
  for _, cat := range all {
    if sameParent(cat.ParentID, parentID) {
      node := ToCategoryResponse(cat)
      node.Subcategories = buildCategoryTree(all, &cat.ID)
      result = append(result, node)
    }
  }
  return result
}


func (cu *CategoryUseCase) GetCategoryTree(ctx context.Context) ([]CategoryResponse, error) {
	categories, err := cu.cr.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return buildCategoryTree(categories, nil), nil
}
