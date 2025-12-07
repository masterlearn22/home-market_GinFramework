package entity

import (
	// "database/sql"
	// "database/sql/driver"
	"github.com/google/uuid"
	"time"
)


type Offer struct {
	ID            uuid.UUID   `db:"id"`
	GiverID       uuid.UUID   `db:"giver_id"`
	SellerID      uuid.UUID   `db:"seller_id"`
	ItemName      string      `db:"item_name"`
	Description   string      `db:"description"`
	ImageURL      string      `db:"image_url"`
	ExpectedPrice float64     `db:"expected_price"`
	AgreedPrice   *float64 	  `db:"agreed_price" json:"agreed_price,omitempty"`
	Condition     string      `db:"condition" json:"condition"`
	Location      string      `db:"location" json:"location"`
	Status        string      `db:"status"` // pending, accepted, rejected, paid
	CreatedAt     time.Time   `db:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at"`
}

type CreateOfferInput struct {
	SellerIDStr     string  `form:"seller_id"`
	ItemName        string  `form:"item_name" binding:"required"`
	Description     string  `form:"description"`
	ExpectedPrice   float64 `form:"expected_price" binding:"min=0"`
	Condition       string  `form:"condition" binding:"required"`
	Location        string  `form:"location"`
	ImageFileHeader string
}

type AcceptOfferInput struct {
	AgreedPrice float64 `json:"agreed_price" binding:"required,min=0"`
}
