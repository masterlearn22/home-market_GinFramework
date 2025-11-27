package service

import (
	"errors"
	"time"
	entity "home-market/internal/domain"
	repo "home-market/internal/repository/postgresql"

	"github.com/google/uuid"
)

var (
	ErrNotSeller      = errors.New("only seller role can create a shop")
	ErrShopExists     = errors.New("user already has a shop")
	ErrNoShopOwned   = errors.New("seller does not own a shop")
)

type ShopService struct {
	shopRepo repo.ShopRepository
}

func NewShopService(shopRepo repo.ShopRepository) *ShopService {
	return &ShopService{shopRepo: shopRepo}
}

func (s *ShopService) CreateShop(userID uuid.UUID, role string, input entity.CreateShopInput) (*entity.Shop, error) {
	// hanya role seller yang boleh buat shop
	if role != "seller" {
		return nil, ErrNotSeller
	}

	// cek apakah user sudah punya toko
	existing, err := s.shopRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrShopExists
	}

	shop := &entity.Shop{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      input.Name,
		Address:   input.Address,
		Description: input.Description,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// simpan toko
	if err := s.shopRepo.CreateShop(shop); err != nil {
		return nil, err
	}

	return shop, nil
}
