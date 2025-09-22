package users

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/ubcent/edge.link/internal/models"
	"golang.org/x/crypto/bcrypt"
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
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.HashedPassword = string(hashedPassword)

	query := `
		INSERT INTO users (tenant_id, email, hashed_password, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	
	err = s.db.QueryRowContext(ctx, query,
		user.TenantID, user.Email, user.HashedPassword, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
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
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

// GetByTenant retrieves users for a specific tenant
func (s *Service) GetByTenant(ctx context.Context, tenantID int) ([]*models.User, error) {
	query := `
		SELECT id, tenant_id, email, hashed_password, role, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for tenant: %w", err)
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
	
	return users, nil
}

// ValidatePassword checks if the provided password matches the user's hashed password
func (s *Service) ValidatePassword(ctx context.Context, user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil
}

// GenerateAPIKeyToken generates a secure API key token
func GenerateAPIKeyToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}