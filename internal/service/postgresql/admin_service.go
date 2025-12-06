// [internal/service/admin_service.go]

package service

import (
	"errors"
	entity "home-market/internal/domain"
	repo "home-market/internal/repository/postgresql"
	"github.com/google/uuid"
)

type AdminService struct {
	userRepo repo.UserRepository // Untuk FR-ADMIN-01, FR-ADMIN-03
	itemRepo repo.ItemRepository // Untuk FR-ADMIN-02 (Moderasi Item)
}

// Pastikan Anda meneruskan kedua repo ini saat inisialisasi
func NewAdminService(userRepo repo.UserRepository, itemRepo repo.ItemRepository) *AdminService {
	return &AdminService{userRepo: userRepo, itemRepo: itemRepo}
}

// FR-ADMIN-01: Manajemen User (List Users)
func (s *AdminService) ListUsers(role string) ([]entity.User, error) {
	if role != "admin" {
		return nil, errors.New("unauthorized: admin access required")
	}
	return s.userRepo.ListAllUsers()
}

// FR-ADMIN-03: Blokir User
func (s *AdminService) BlockUser(adminRole string, targetUserID uuid.UUID, isActive bool) error {
	if adminRole != "admin" {
		return errors.New("unauthorized: admin access required")
	}
	// FR-ADMIN-03: Set is_active = false [cite: 478]
	return s.userRepo.UpdateUserStatus(targetUserID, isActive)
}

// FR-ADMIN-02: Moderasi Barang (Set status='inactive')
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