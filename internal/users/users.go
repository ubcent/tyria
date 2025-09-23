package users

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/ubcent/edge.link/internal/auth"
	"github.com/ubcent/edge.link/internal/models"
)

// Service provides user management functionality
type Service struct {
	db *sql.DB
}

// NewService creates a new user service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Create creates a new user with hashed password
func (s *Service) Create(ctx context.Context, user *models.User, password string) error {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (tenant_id, email, hashed_password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err = s.db.QueryRowContext(ctx, query,
		user.TenantID, user.Email, hashedPassword, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.HashedPassword = hashedPassword
	return nil
}

// Authenticate validates user credentials and returns the user if valid
func (s *Service) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	valid, err := auth.VerifyPassword(password, user.HashedPassword)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("authentication failed: invalid credentials")
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (s *Service) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, tenant_id, email, hashed_password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.HashedPassword,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (s *Service) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, tenant_id, email, hashed_password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.HashedPassword,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByTenant retrieves users for a specific tenant
func (s *Service) GetByTenant(ctx context.Context, tenantID int) ([]*models.User, error) {
	query := `
		SELECT id, tenant_id, email, hashed_password, role, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.TenantID, &user.Email, &user.HashedPassword,
			&user.Role, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// UpdatePassword updates a user's password
func (s *Service) UpdatePassword(ctx context.Context, userID int, newPassword string) error {
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		UPDATE users 
		SET hashed_password = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateRole updates a user's role
func (s *Service) UpdateRole(ctx context.Context, userID int, role string) error {
	if !auth.IsValidRole(role) {
		return fmt.Errorf("invalid role: %s", role)
	}

	query := `
		UPDATE users 
		SET role = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.db.ExecContext(ctx, query, role, userID)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete deletes a user
func (s *Service) Delete(ctx context.Context, userID int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ValidatePassword checks if the provided password matches the user's hashed password
func (s *Service) ValidatePassword(ctx context.Context, user *models.User, password string) bool {
	valid, err := auth.VerifyPassword(password, user.HashedPassword)
	return err == nil && valid
}

// GenerateAPIKeyToken generates a secure API key token
func GenerateAPIKeyToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}