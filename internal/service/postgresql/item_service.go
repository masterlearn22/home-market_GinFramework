package service

import (
	"errors"
	"time"

	entity "home-market/internal/domain"
	repo "home-market/internal/repository/postgresql"

	"github.com/google/uuid"
)

var (
	ErrInvalidStock      = errors.New("stock must be >= 0")
	ErrInvalidPrice      = errors.New("price must be >= 0")
	ErrCategoryNotOwned  = errors.New("category does not belong to seller's shop")
	ErrNotGiver = errors.New("access denied: only giver role is allowed")
)

type ItemService struct {
	itemRepo repo.ItemRepository
}

func NewItemService(itemRepo repo.ItemRepository) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
	}
}

func (s *ItemService) CreateItem(userID uuid.UUID, role string, input entity.CreateItemInput, imageURLs []string) (*entity.Item, []entity.ItemImage, error) {

	if role != "seller" {
		return nil, nil, ErrNotSeller
	}

	shop, err := s.itemRepo.GetShopByUserID(userID)
	if err != nil {
		return nil, nil, err
	}
	if shop == nil {
		return nil, nil, ErrNoShopOwned
	}

	// Validasi kategori milik shop
	owned, err := s.itemRepo.IsCategoryOwnedByShop(input.CategoryID, shop.ID)
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

	// Simpan item
	if err := s.itemRepo.CreateItem(item); err != nil {
		return nil, nil, err
	}

	// Simpan gambar
	var images []entity.ItemImage
	for _, url := range imageURLs {
		img := entity.ItemImage{
			ID:       uuid.New(),
			ItemID:   item.ID,
			ImageURL: url,
			CreatedAt: time.Now(),
		}

		if err := s.itemRepo.CreateItemImage(&img); err != nil {
			return item, nil, err
		}

		images = append(images, img)
	}

	return item, images, nil
}

func (s *ItemService) UpdateItem(userID uuid.UUID, itemID uuid.UUID, input entity.UpdateItemInput) (*entity.Item, error) {
    // 1. Cek apakah Item ada
    item, err := s.itemRepo.GetItemByID(itemID)
    if err != nil {
        return nil, err
    }
    if item == nil {
        return nil, errors.New("item not found")
    }

    // 2. Cek Shop milik User
    shop, err := s.itemRepo.GetShopByUserID(userID)
    if err != nil {
        return nil, err
    }
    if shop == nil {
        return nil, errors.New("you do not have a shop")
    }

    // 3. Validasi Kepemilikan: Item.ShopID harus sama dengan Shop.ID milik user
    if item.ShopID != shop.ID {
        return nil, errors.New("unauthorized: this item does not belong to your shop")
    }

    // 4. Update Field (FR-SELLER-04 & FR-SELLER-06)
    item.Name = input.Name
    item.Description = input.Description
    item.Price = input.Price
    item.Stock = input.Stock
    item.Condition = input.Condition
    
    // Jika user mengirim status (misal mau re-activate), pakai itu. Jika kosong, biarkan yang lama.
    if input.Status != "" {
        item.Status = input.Status
    }

    // 5. Simpan ke DB
    if err := s.itemRepo.UpdateItem(item); err != nil {
        return nil, err
    }

    return item, nil
}


func (s *ItemService) DeleteItem(userID uuid.UUID, itemID uuid.UUID) error {
    // 1. Cek Item & Shop (Logic sama seperti update)
    item, err := s.itemRepo.GetItemByID(itemID)
    if err != nil {
        return err
    }
    if item == nil {
        return errors.New("item not found")
    }

    shop, err := s.itemRepo.GetShopByUserID(userID)
    if err != nil {
        return err
    }
    if shop == nil || item.ShopID != shop.ID {
        return errors.New("unauthorized")
    }

   
    item.Status = "inactive"

    // 3. Simpan perubahan
    return s.itemRepo.UpdateItem(item)
}

// FR-GIVER-01 & FR-GIVER-02: Membuat Penawaran Barang
func (s *ItemService) CreateOffer(userID uuid.UUID, role string, input entity.CreateOfferInput, imageURL string) (*entity.Offer, error) {
    if role != "giver" {
        return nil, ErrNotGiver // Validasi FR-GIVER-01
    }

    var sellerID uuid.UUID
    if input.SellerIDStr != "" {
        id, err := uuid.Parse(input.SellerIDStr)
        if err != nil {
            return nil, errors.New("invalid seller_id format")
        }
        // Opsional: Cek apakah SellerID valid dan memiliki toko aktif (tambahan validasi)
        sellerID = id
    }
    
    // Validasi dasar
    if input.ExpectedPrice < 0 {
        return nil, errors.New("expected price cannot be negative")
    }

    offer := &entity.Offer{
        ID:             uuid.New(),
        GiverID:        userID,
        SellerID:       sellerID,
        ItemName:       input.ItemName,
        Description:    input.Description,
        ImageURL:       imageURL, // URL dari file yang di-upload (FR-GIVER-02)
        ExpectedPrice:  input.ExpectedPrice,
        Condition:      input.Condition,
        Location:       input.Location,
        Status:         "pending", // Status awal selalu pending (FR-GIVER-01)
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }

    if err := s.itemRepo.CreateOffer(offer); err != nil {
        return nil, err
    }
    
    // Opsional: Trigger notifikasi ke Seller jika SellerID ada (FR-NOTIF-01)
    
    return offer, nil
}

// FR-GIVER-03: Melihat Status Penawaran
func (s *ItemService) GetMyOffers(userID uuid.UUID, role string) ([]entity.Offer, error) {
    if role != "giver" {
        return nil, ErrNotGiver
    }
    
    return s.itemRepo.GetOffersByGiverID(userID)
}