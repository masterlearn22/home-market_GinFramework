package entity

import (
	"time"
	"database/sql"
	"github.com/google/uuid"
)

type Offer struct {
	ID          uuid.UUID `db:"id"`
	GiverID     uuid.UUID `db:"giver_id"`
	SellerID    uuid.UUID `db:"seller_id"`
	ItemName    string    `db:"item_name"`
	Description string    `db:"description"`
	ImageURL    string    `db:"image_url"`
	ExpectedPrice float64 `db:"expected_price"`
	AgreedPrice   sql.NullFloat64 `db:"agreed_price"`
	Condition      string    `db:"condition" json:"condition"`
    Location       string    `db:"location" json:"location"`
	Status      string    `db:"status"`  // pending, accepted, rejected, paid
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type CreateOfferInput struct {
    // SellerID opsional, bisa string untuk parsing atau pointer
	SellerIDStr     string  `form:"seller_id"` // Opsional untuk Open Offer
	ItemName        string  `form:"item_name" binding:"required"`
	Description     string  `form:"description"`
	ExpectedPrice   float64 `form:"expected_price" binding:"min=0"`
	Condition       string  `form:"condition" binding:"required"`
    Location        string  `form:"location"`
	ImageFileHeader string  // Placeholder untuk nama file yang di-upload
}
