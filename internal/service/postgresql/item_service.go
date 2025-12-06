package service

import (
	"errors"
	"github.com/google/uuid"
	entity "home-market/internal/domain"
	repo "home-market/internal/repository/postgresql"
	"time"
)

var (
	ErrInvalidStock     = errors.New("stock must be >= 0")
	ErrInvalidPrice     = errors.New("price must be >= 0")
	ErrCategoryNotOwned = errors.New("category does not belong to seller's shop")
	ValidOrderStatuses = map[string]bool{
	"pending": true, "paid": true, "processing": true,
	"shipped": true, "completed": true, "cancelled": true,
	}
)

type ItemService struct {
	itemRepo repo.ItemRepository
	shopRepo repo.ShopRepository
	orderRepo repo.OrderRepository
}

func NewItemService(itemRepo repo.ItemRepository, shopRepo repo.ShopRepository, orderRepo repo.OrderRepository) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
		shopRepo: shopRepo,
		orderRepo: orderRepo,
	}
}

// @Summary      Create New Item
// @Description  Allows a Seller to create a new item within their shop. Requires multipart/form-data for input and image upload.
// @Tags         Seller/Items
// @Accept       mpfd
// @Produce      json
// @Security     ApiKeyAuth
// @Param        name formData string true "Item Name"
// @Param        description formData string false "Item Description"
// @Param        price formData number true "Item Price"
// @Param        stock formData integer true "Initial Stock"
// @Param        condition formData string false "Item Condition"
// @Param        category_id formData string true "Category ID (UUID) owned by the shop"
// @Param        images formData file true "Item Images"
// @Success      201  {object}  map[string]interface{} "Returns created item and image URLs"
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{} "Forbidden (not seller)"
// @Failure      500  {object}  map[string]interface{}
// @Router       /items [post]
func (s *ItemService) CreateItem(userID uuid.UUID, role string, input entity.CreateItemInput, imageURLs []string) (*entity.Item, []entity.ItemImage, error) {

	if role != "seller" {
		return nil, nil, ErrNotSeller
	}

	shop, err := s.shopRepo.GetByUserID(userID)
	if err != nil {
		return nil, nil, err
	}
	if shop == nil {
		return nil, nil, ErrNoShopOwned
	}

	
	owned, err := s.shopRepo.IsCategoryOwnedByShop(input.CategoryID, shop.ID)
	if err != nil {
		return nil, nil, err
	}
	if !owned {
		return nil, nil, ErrCategoryNotOwned
	}

	if input.Stock < 0 {
		return nil, nil, ErrInvalidStock
	}
	if input.Price < 0 {
		return nil, nil, ErrInvalidPrice
	}

	item := &entity.Item{
		ID:          uuid.New(),
		ShopID:      shop.ID,
		CategoryID:  input.CategoryID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		Condition:   input.Condition,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}


	if err := s.itemRepo.CreateItem(item); err != nil {
		return nil, nil, err
	}


	var images []entity.ItemImage
	for _, url := range imageURLs {
		img := entity.ItemImage{
			ID:        uuid.New(),
			ItemID:    item.ID,
			ImageURL:  url,
			CreatedAt: time.Now(),
		}

		if err := s.itemRepo.CreateItemImage(&img); err != nil {
			return item, nil, err
		}

		images = append(images, img)
	}

	return item, images, nil
}

// @Summary      Update Item Details
// @Description  Allows a Seller to update item fields (name, price, stock, status, etc.) for an item they own.
// @Tags         Seller/Items
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Item ID to update"
// @Param        input body entity.UpdateItemInput true "Updated item details"
// @Success      200  {object}  entity.Item "Returns the updated item"
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{} "Unauthorized (not owner)"
// @Failure      404  {object}  map[string]interface{} "Item not found"
// @Failure      500  {object}  map[string]interface{}
// @Router       /items/{id} [put]
func (s *ItemService) UpdateItem(userID uuid.UUID, itemID uuid.UUID, input entity.UpdateItemInput) (*entity.Item, error) {
	
	item, err := s.itemRepo.GetItemByID(itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}

	shop, err := s.shopRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if shop == nil {
		return nil, errors.New("you do not have a shop")
	}

	if item.ShopID != shop.ID {
		return nil, errors.New("unauthorized: this item does not belong to your shop")
	}


	item.Name = input.Name
	item.Description = input.Description
	item.Price = input.Price
	item.Stock = input.Stock
	item.Condition = input.Condition

	
	if input.Status != "" {
		item.Status = input.Status
	}

	if err := s.itemRepo.UpdateItem(item); err != nil {
		return nil, err
	}

	return item, nil
}

// @Summary      Archive/Delete Item (Soft Delete)
// @Description  Sets the status of an item to 'inactive'. Only the item owner can perform this action.
// @Tags         Seller/Items
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Item ID to delete"
// @Success      200  {object}  map[string]interface{} "Item archived/deleted successfully"
// @Failure      403  {object}  map[string]interface{} "Unauthorized"
// @Failure      404  {object}  map[string]interface{} "Item not found"
// @Failure      500  {object}  map[string]interface{}
// @Router       /items/{id} [delete]
func (s *ItemService) DeleteItem(userID uuid.UUID, itemID uuid.UUID) error {
	item, err := s.itemRepo.GetItemByID(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("item not found")
	}

	shop, err := s.shopRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if shop == nil || item.ShopID != shop.ID {
		return errors.New("unauthorized")
	}

	item.Status = "inactive"
	return s.itemRepo.UpdateItem(item)
}

// @Summary      Get Item Detail (Marketplace View)
// @Description  Retrieves detailed information for a single item, ensuring it is active and available.
// @Tags         Marketplace
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Item ID to retrieve"
// @Success      200  {object}  entity.Item
// @Failure      404  {object}  map[string]interface{} "Item not found or inactive"
// @Failure      500  {object}  map[string]interface{}
// @Router       /market/items/{id} [get]
func (s *ItemService) GetItemDetail(itemID uuid.UUID) (*entity.Item, error) {
	item, err := s.orderRepo.GetItemForOrder(itemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.Status != "active" {
		return nil, errors.New("item not found or inactive")
	}
	return item, nil
}



