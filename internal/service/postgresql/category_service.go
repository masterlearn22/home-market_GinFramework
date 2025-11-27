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

func (s *CategoryService) CreateCategory(userID uuid.UUID, role string, input entity.CreateCategoryInput) (*entity.Category, error) {

	if role != "seller" {
		return nil, ErrNotSeller
	}

	// pastikan seller punya shop
	shop, err := s.categoryRepo.GetShopByUserID(userID)
	if err != nil {
		return nil, err
	}

	if shop == nil {
		return nil, ErrNoShopOwned
	}

	//Pengecekan apakah nama kategori sudah ada atau belum
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
