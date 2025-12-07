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

// @Summary      Create Seller Shop
// @Description  Allows a registered Seller to create their shop. A user can only own one shop.
// @Tags         Shop
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        input body entity.CreateShopInput true "Shop details (Name, Address, Description)"
// @Success      201  {object}  entity.Shop
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{} "Forbidden (not seller)"
// @Failure      409  {object}  map[string]interface{} "Conflict (shop already exists)"
// @Failure      500  {object}  map[string]interface{}
// @Router       /shops [post]
func (s *ShopService) CreateShop(userID uuid.UUID, role string, input entity.CreateShopInput) (*entity.Shop, error) {
	if role != "seller" {
		return nil, ErrNotSeller
	}

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

	if err := s.shopRepo.CreateShop(shop); err != nil {
		return nil, err
	}

	return shop, nil
}
