package service

import (
	"time"
	entity "home-market/internal/domain"
	repo "home-market/internal/repository/postgresql"
	"errors"
	"github.com/google/uuid"
)

var (
	ErrCategoryExists = errors.New("category name already exists in this shop")
)

type CategoryService struct {
	categoryRepo repo.CategoryRepository
}

func NewCategoryService(categoryRepo repo.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// @Summary      Create New Shop Category
// @Description  Allows a logged-in seller to create a new category unique to their shop.
// @Tags         Category
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        input body entity.CreateCategoryInput true "Category details (Name)"
// @Success      201  {object}  entity.Category
// @Failure      400  {object}  map[string]interface{} "Invalid input or missing shop"
// @Failure      403  {object}  map[string]interface{} "Forbidden (not seller)"
// @Failure      409  {object}  map[string]interface{} "Conflict (category name already exists)"
// @Failure      500  {object}  map[string]interface{} "Internal server error"
// @Router       /categories [post]
func (s *CategoryService) CreateCategory(userID uuid.UUID, role string, input entity.CreateCategoryInput) (*entity.Category, error) {

	if role != "seller" {
		return nil, ErrNotSeller
	}

	shop, err := s.categoryRepo.GetShopByUserID(userID)
	if err != nil {
		return nil, err
	}

	if shop == nil {
		return nil, ErrNoShopOwned
	}

	exists, err := s.categoryRepo.ExistsByName(shop.ID, input.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCategoryExists
	}

	category := &entity.Category{
		ID:        uuid.New(),
		ShopID:    shop.ID,
		Name:      input.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.categoryRepo.CreateCategory(category)
	if err != nil {
		return nil, err
	}

	return category, nil
}
