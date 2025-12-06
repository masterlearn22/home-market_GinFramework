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
	ListAllUsers() ([]entity.User, error) // FR-ADMIN-01
    UpdateUserStatus(userID uuid.UUID, isActive bool) error // FR-ADMIN-03
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

func (r *userRepository) ListAllUsers() ([]entity.User, error) {
    // 1. Deklarasikan slice untuk menampung hasil (FIX: undefined users)
    var users []entity.User 

    query := `
        SELECT id, name, email is_active, created_at, updated_at
        FROM users
    `
    // Eksekusi query
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // 2. Loop melalui hasil query
    for rows.Next() {
        var user entity.User // Asumsi entity.User memiliki semua field di atas
        
        // Asumsi struct entity.User mencakup field is_active dll.
        // Anda perlu menyesuaikan Scan() ini agar sesuai dengan struct entity.User Anda.
        err := rows.Scan(
            &user.ID, &user.FullName, &user.Email, &user.IsActive, 
            &user.CreatedAt, &user.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    // 3. Cek error saat iterasi selesai
    if err = rows.Err(); err != nil {
        return nil, err
    }

    // Mengembalikan slice users (yang mungkin kosong jika tidak ada data)
    return users, nil
}

func (r *userRepository) UpdateUserStatus(userID uuid.UUID, isActive bool) error {
    query := `UPDATE users SET is_active = $1 WHERE id = $2`
    _, err := r.db.Exec(query, isActive, userID)
    return err
}