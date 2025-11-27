package entity
import (
	"time"
	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID `db:"id"`
	ShopID    uuid.UUID `json:"shopId"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CreateCategoryInput struct {
	Name string `json:"name" binding:"required"`
}
