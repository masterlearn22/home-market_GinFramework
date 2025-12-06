package service

import (
	"errors"
	entity "home-market/internal/domain"
	repo "home-market/internal/repository/postgresql"
	"github.com/google/uuid"
)

type AdminService struct {
	userRepo repo.UserRepository 
	itemRepo repo.ItemRepository 
}

func NewAdminService(userRepo repo.UserRepository, itemRepo repo.ItemRepository) *AdminService {
	return &AdminService{userRepo: userRepo, itemRepo: itemRepo}
}

// @Summary      Get List of All Users
// @Description  Retrieves a list of all users in the system (Admin access only).
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   entity.User
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /admin/users [get]
func (s *AdminService) ListUsers(role string) ([]entity.User, error) {
	if role != "admin" {
		return nil, errors.New("unauthorized: admin access required")
	}
	return s.userRepo.ListAllUsers()
}

// @Summary      Block or Unblock User Account
// @Description  Sets the 'is_active' status of a target user (Admin access only).
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Target User ID"
// @Param        input body entity.UpdateUserStatusInput true "Status update (is_active: true/false)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /admin/users/{id}/status [patch]
func (s *AdminService) BlockUser(adminRole string, targetUserID uuid.UUID, isActive bool) error {
	if adminRole != "admin" {
		return errors.New("unauthorized: admin access required")
	}
	// FR-ADMIN-03: Set is_active = false [cite: 478]
	return s.userRepo.UpdateUserStatus(targetUserID, isActive)
}

// @Summary      Moderate Item (Set Inactive)
// @Description  Sets the status of a specific item to 'inactive' due to policy violation (Admin access only).
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Item ID to moderate"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /admin/items/{id}/moderate [patch]
func (s *AdminService) ModerateItem(adminRole string, itemID uuid.UUID) error {
	if adminRole != "admin" {
		return errors.New("unauthorized: admin access required")
	}
	
	item, err := s.itemRepo.GetItemByID(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("item not found")
	}

	// FR-ADMIN-02: Ubah status item menjadi 'inactive' [cite: 473]
	item.Status = "inactive" 
	return s.itemRepo.UpdateItem(item) 
}