package service

import (
	// "database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	entity "home-market/internal/domain"
	mongorepo "home-market/internal/repository/mongodb"
	repo "home-market/internal/repository/postgresql"
	"log"
	"time"
)

var (
	ErrNotGiver         = errors.New("access denied: only giver role is allowed")
	ErrOfferNotFound    = errors.New("offer not found")
	ErrOfferStatus      = errors.New("offer is not in pending status")
	ErrNotSellerOrOwner = errors.New("unauthorized: access denied or you are not the owner")
)

type OfferService struct {
	offerRepo repo.OfferRepository
	itemRepo  repo.ItemRepository
	shopRepo  repo.ShopRepository
	logRepo   mongorepo.LogRepository
}

func NewOfferService(offerRepo repo.OfferRepository, itemRepo repo.ItemRepository, shopRepo repo.ShopRepository, logRepo mongorepo.LogRepository) *OfferService {
	return &OfferService{
		offerRepo: offerRepo,
		itemRepo:  itemRepo,
		shopRepo:  shopRepo,
		logRepo:   logRepo,
	}
}

func (s *OfferService) checkSellerOwnership(userID uuid.UUID) (*entity.Shop, error) {
	shop, err := s.shopRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	return shop, nil
}

func (s *OfferService) createDraftItemFromOffer(offer *entity.Offer, shopID uuid.UUID) *entity.Item {
	var price float64
    if offer.AgreedPrice != nil {
        price = *offer.AgreedPrice
    } else {
        price = offer.ExpectedPrice 
    }
	return &entity.Item{
		ID:          uuid.New(),
		ShopID:      shopID,
		CategoryID:  uuid.Nil,
		Name:        offer.ItemName,
		Description: fmt.Sprintf("Draft dari Penawaran: %s. Kondisi: %s. Lokasi Awal: %s.", offer.Description, offer.Condition, offer.Location),
		Price:       price,
		Stock:       1,
		Condition:   offer.Condition,
		Status:      "draft",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (s *OfferService) createAndSaveNotification(userID uuid.UUID, title string, message string, notiType string, relatedID uuid.UUID) {
	noti := &entity.Notification{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		Title:     title,
		Message:   message,
		Type:      notiType,
		RelatedID: relatedID,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := s.logRepo.SaveNotification(noti); err != nil {
		log.Printf("Warning: failed to save notification for user %s: %v", userID.String(), err)
	}
}

// @Summary      Create a New Offer
// @Description  Allows a Giver to create an offer for an item, optionally targeting a specific Seller. Requires multipart/form-data.
// @Tags         Offers
// @Accept       mpfd
// @Produce      json
// @Security     ApiKeyAuth
// @Param        item_name formData string true "Name of the item being offered"
// @Param        description formData string false "Description of the item"
// @Param        expected_price formData number true "Expected price in IDR"
// @Param        condition formData string true "Item condition (e.g., new, used)"
// @Param        location formData string true "Giver's location"
// @Param        seller_id formData string false "Optional Seller ID to target"
// @Param        images formData file true "Item image file"
// @Success      201  {object}  entity.Offer
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /offers [post]
func (s *OfferService) CreateOffer(userID uuid.UUID, role string, input entity.CreateOfferInput, imageURL string) (*entity.Offer, error) {
	if role != "giver" {
		return nil, ErrNotGiver
	}

	var sellerID uuid.UUID
	if input.SellerIDStr != "" {
		id, err := uuid.Parse(input.SellerIDStr)
		if err != nil {
			return nil, errors.New("invalid seller_id format")
		}
		sellerID = id
	}

	if input.ExpectedPrice < 0 {
		return nil, errors.New("expected price cannot be negative")
	}

	offer := &entity.Offer{
		ID:            uuid.New(),
		GiverID:       userID,
		SellerID:      sellerID,
		ItemName:      input.ItemName,
		Description:   input.Description,
		ImageURL:      imageURL,
		ExpectedPrice: input.ExpectedPrice,
		Condition:     input.Condition,
		Location:      input.Location,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.offerRepo.CreateOffer(offer); err != nil {
		return nil, err
	}

	if offer.SellerID != uuid.Nil {
		s.createAndSaveNotification(offer.SellerID, "Penawaran Baru Masuk", fmt.Sprintf("Anda menerima penawaran dari Giver untuk barang '%s'.", offer.ItemName), "offer", offer.ID)
	}

	return offer, nil
}

// @Summary      View My Outgoing Offers
// @Description  Allows the Giver to view the status of all offers they have created.
// @Tags         Offers
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   entity.Offer
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /offers/my [get]
func (s *OfferService) GetMyOffers(userID uuid.UUID, role string) ([]entity.Offer, error) {
	if role != "giver" {
		return nil, ErrNotGiver
	}
	return s.offerRepo.GetOffersByGiverID(userID)
}

// @Summary      View Seller Offer Inbox
// @Description  Allows a Seller to view pending offers directed to them or general open offers.
// @Tags         Offers
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   entity.Offer
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /offers/inbox [get]
func (s *OfferService) GetOffersToSeller(userID uuid.UUID, role string) ([]entity.Offer, error) {
	if role != "seller" {
		return nil, errors.New("access denied: only seller can view offers")
	}

	shop, err := s.shopRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if shop == nil {
		return nil, ErrNoShopOwned
	}
	return s.offerRepo.GetOffersBySellerID(userID)
}

// @Summary      Accept Offer and Create Item Draft
// @Description  Allows the Seller to accept a pending offer, setting the agreed price and generating a draft item for their shop.
// @Tags         Offers
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{} "Returns the updated offer and the newly created item draft"
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{} "Unauthorized or not owner"
// @Failure      404  {object}  map[string]interface{} "Offer not found"
// @Failure      409  {object}  map[string]interface{} "Offer status is not pending"
// @Failure      500  {object}  map[string]interface{}
// @Router       /offers/{id}/accept [post]
func (s *OfferService) AcceptOffer(userID uuid.UUID, offerID uuid.UUID, input entity.AcceptOfferInput) (*entity.Offer, *entity.Item, error) {
	var shop *entity.Shop

	shop, err := s.checkSellerOwnership(userID)
	if err != nil {
		return nil, nil, err
	}
	if shop == nil {
		return nil, nil, ErrNoShopOwned
	}

	offer, err := s.offerRepo.GetOfferByID(offerID)
	if err != nil {
		return nil, nil, err
	}
	if offer == nil {
		return nil, nil, ErrOfferNotFound
	}

	if offer.SellerID != uuid.Nil && offer.SellerID != userID {
		return nil, nil, ErrNotSellerOrOwner
	}

	if offer.Status != "pending" {
		return nil, nil, ErrOfferStatus
	}

	oldStatus := offer.Status
	offer.Status = "accepted"
	price := input.AgreedPrice
    offer.AgreedPrice = &price

	if err := s.offerRepo.UpdateOffer(offer); err != nil {
		return nil, nil, err
	}
	draftItem := s.createDraftItemFromOffer(offer, shop.ID)
	if err := s.itemRepo.CreateItem(draftItem); err != nil { 
		return offer, nil, errors.New("offer accepted, but failed to create draft item")
	}

	
	history := &entity.HistoryStatus{
		ID:          primitive.NewObjectID(),
		RelatedID:   offerID.String(),
		RelatedType: "offer",
		OldStatus:   oldStatus,
		NewStatus:   "accepted",
		ChangedBy:   userID.String(),
		Timestamp:   time.Now(),
	}
	if err := s.logRepo.SaveHistoryStatus(history); err != nil {
		log.Printf("Warning: failed to save history status for offer %s: %v", offerID.String(), err)
	}

	return offer, draftItem, nil
}

// @Summary      Reject Offer
// @Description  Allows the Seller to reject a pending offer, setting the status to 'rejected'.
// @Tags         Offers
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Offer ID to reject"
// @Success      200  {object}  entity.Offer
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{} "Unauthorized or not owner"
// @Failure      404  {object}  map[string]interface{} "Offer not found"
// @Failure      409  {object}  map[string]interface{} "Offer status is not pending"
// @Failure      500  {object}  map[string]interface{}
// @Router       /offers/{id}/reject [post]
func (s *OfferService) RejectOffer(userID uuid.UUID, offerID uuid.UUID) (*entity.Offer, error) {
	if shop, err := s.checkSellerOwnership(userID); err != nil {
		return nil, err
	} else if shop == nil {
		return nil, ErrNoShopOwned
	}

	offer, err := s.offerRepo.GetOfferByID(offerID)
	if err != nil {
		return nil, err
	}
	if offer == nil {
		return nil, ErrOfferNotFound
	}
	if offer.Status != "pending" {
		return nil, ErrOfferStatus
	}

	oldStatus := offer.Status

	offer.Status = "rejected"

	if err := s.offerRepo.UpdateOffer(offer); err != nil {
		return nil, err
	}

	history := &entity.HistoryStatus{
		ID:          primitive.NewObjectID(),
		RelatedID:   offerID.String(),
		RelatedType: "offer",
		OldStatus:   oldStatus,
		NewStatus:   "rejected",
		ChangedBy:   userID.String(),
		Timestamp:   time.Now(),
	}
	if err := s.logRepo.SaveHistoryStatus(history); err != nil {
		log.Printf("Warning: failed to save history status for offer %s: %v", offerID.String(), err)
	}

	return offer, nil
}
