package repository

import (
	"database/sql"
	"errors"

	entity "home-market/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByUsername(username string) (*entity.User, string, error)
	GetPermissionsByRoleID(roleID uuid.UUID) ([]string, error)
	GetByID(id uuid.UUID) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	CreateUser(user *entity.User) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByUsername(username string) (*entity.User, string, error) {
	var user entity.User
	var roleName string

	query := `
		SELECT 
			u.id, u.username, u.email, u.password_hash, 
			u.full_name, u.role_id, u.is_active,
			TRIM(r.name) AS roleName
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1
	`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
		&roleName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("user not found")
		}
		return nil, "", err
	}

	return &user, roleName, nil
}

func (r *userRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]string, error) {
	query := `
		SELECT p.name 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`

	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := []string{}
	for rows.Next() {
		var permName string
		if err := rows.Scan(&permName); err != nil {
			return nil, err
		}
		permissions = append(permissions, permName)
	}

	return permissions, nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*entity.User, error) {
	var user entity.User

	query := `
		SELECT id, username, email, full_name, role_id, is_active
		FROM users
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*entity.User, error) {
	var user entity.User

	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active
		FROM users
		WHERE email = $1
	`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // <--- lebih singkat untuk mengecek "email available"
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) CreateUser(user *entity.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`

	_, err := r.db.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.RoleID,
		user.IsActive,
	)

	return err
}
