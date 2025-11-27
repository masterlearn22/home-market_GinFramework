package repository

import (
	"database/sql"
	entity "home-market/internal/domain"

	"github.com/google/uuid"
)




type CategoryRepository interface {
	CreateCategory(c *entity.Category) error
	GetShopByUserID(userID uuid.UUID) (*entity.Shop, error)
	ExistsByName(shopID uuid.UUID, name string) (bool, error)
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Ambil toko milik seller
func (r *categoryRepository) GetShopByUserID(userID uuid.UUID) (*entity.Shop, error) {
	var shop entity.Shop

	query := `
		SELECT id, name, description, address, created_at, updated_at
		FROM shops
		WHERE user_id = $1
	`

	err := r.db.QueryRow(query, userID).Scan(
		&shop.ID,
		&shop.Name,
		&shop.Description,
		&shop.Address,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // tidak punya toko
	}

	if err != nil {
		return nil, err
	}

	return &shop, nil
}

// Insert kategori
func (r *categoryRepository) CreateCategory(c *entity.Category) error {
	query := `
		INSERT INTO categories (id, shop_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`
	_, err := r.db.Exec(query,
		c.ID,
		c.ShopID,
		c.Name,
	)

	return err
}

func (r *categoryRepository) ExistsByName(shopID uuid.UUID, name string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM categories
			WHERE shop_id = $1 AND LOWER(name) = LOWER($2)
		)
	`

	err := r.db.QueryRow(query, shopID, name).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

