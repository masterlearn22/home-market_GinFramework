package entity

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` 
	FullName     string    `json:"full_name" db:"full_name"`
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Struct untuk FR-ADMIN-03
type UpdateUserStatusInput struct {
    IsActive bool `json:"is_active" binding:"required"`
}