package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"crm/models"
)

// UserRepository handles all user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, role, status, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, role, status, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// ListUsers returns all users with optional filters
func (r *UserRepository) ListUsers(page, pageSize int) ([]*models.User, int, error) {
	offset := (page - 1) * pageSize
	rows, err := r.db.Query(`
		SELECT id, name, email, role, status, created_at, updated_at
		FROM users WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	var total int
	r.db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&total)
	return users, total, nil
}

// Create creates a new user
func (r *UserRepository) Create(req *models.CreateUserRequest, passwordHash string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, password_hash, role, status)
		VALUES ($1, $2, $3, $4, 'active')
		RETURNING id, name, email, role, status, created_at, updated_at
	`, req.Name, req.Email, passwordHash, req.Role).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

// Update updates a user's details
func (r *UserRepository) Update(id string, req *models.UpdateUserRequest) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(`
		UPDATE users SET name=$1, email=$2, role=$3, status=$4, updated_at=NOW()
		WHERE id=$5 AND deleted_at IS NULL
		RETURNING id, name, email, role, status, created_at, updated_at
	`, req.Name, req.Email, req.Role, req.Status, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

// UpdatePassword updates a user's password hash
func (r *UserRepository) UpdatePassword(id, passwordHash string) error {
	_, err := r.db.Exec(`
		UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL
	`, passwordHash, id)
	return err
}

// SoftDelete marks a user as deleted
func (r *UserRepository) SoftDelete(id string) error {
	_, err := r.db.Exec(`
		UPDATE users SET deleted_at=$1 WHERE id=$2 AND deleted_at IS NULL
	`, time.Now(), id)
	return err
}

// ListUsersSimple returns a simple list of users for dropdowns
func (r *UserRepository) ListUsersSimple() ([]*models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, name, email, role FROM users WHERE deleted_at IS NULL AND status='active' ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE email=$1 AND deleted_at IS NULL", email).Scan(&count)
	return count > 0, err
}

// ExistsByEmailExcluding checks if a user exists with the given email excluding an id
func (r *UserRepository) ExistsByEmailExcluding(email, excludeID string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE email=$1 AND id!=$2 AND deleted_at IS NULL", email, excludeID).Scan(&count)
	return count > 0, err
}

// CountUsers returns total user count
func (r *UserRepository) CountUsers() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&count)
	return count, err
}

// UpdateLastLogin is a no-op placeholder (could track in a separate table)
func (r *UserRepository) UpdateLastLogin(id string) error {
	_, err := r.db.Exec("UPDATE users SET updated_at=NOW() WHERE id=$1", id)
	return err
}

// scanUser is a helper to scan into a user struct
func scanUser(row *sql.Row) (*models.User, error) {
	u := &models.User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}
