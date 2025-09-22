package apikeys

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/ubcent/edge.link/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Service provides API key management functionality
type Service struct {
	db *sql.DB
}

// NewService creates a new API key service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Create creates a new API key
func (s *Service) Create(ctx context.Context, apiKey *models.APIKey) (string, error) {
	// Generate a secure API key
	fullKey, err := s.generateAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Extract prefix (first 20 characters for display)
	apiKey.Prefix = fullKey[:20] + "..."

	// Hash the full key for storage
	hashedKey, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash API key: %w", err)
	}
	apiKey.Hash = string(hashedKey)

	query := `
		INSERT INTO api_keys (tenant_id, name, prefix, hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	
	err = s.db.QueryRowContext(ctx, query,
		apiKey.TenantID, apiKey.Name, apiKey.Prefix, apiKey.Hash,
	).Scan(&apiKey.ID, &apiKey.CreatedAt, &apiKey.UpdatedAt)
	
	if err != nil {
		return "", fmt.Errorf("failed to create API key: %w", err)
	}
	
	return fullKey, nil
}

// GetByTenant retrieves API keys for a specific tenant
func (s *Service) GetByTenant(ctx context.Context, tenantID int) ([]*models.APIKey, error) {
	query := `
		SELECT id, tenant_id, name, prefix, hash, last_used_at, created_at, updated_at
		FROM api_keys
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys for tenant: %w", err)
	}
	defer rows.Close()
	
	var keys []*models.APIKey
	for rows.Next() {
		key := &models.APIKey{}
		err := rows.Scan(
			&key.ID, &key.TenantID, &key.Name, &key.Prefix, &key.Hash,
			&key.LastUsedAt, &key.CreatedAt, &key.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}
	
	return keys, nil
}

// ValidateKey validates an API key and returns the associated key info
func (s *Service) ValidateKey(ctx context.Context, keyValue string) (*models.APIKey, error) {
	// Extract prefix from the key
	var prefix string
	if len(keyValue) >= 20 {
		prefix = keyValue[:20] + "..."
	}

	query := `
		SELECT id, tenant_id, name, prefix, hash, last_used_at, created_at, updated_at
		FROM api_keys
		WHERE prefix = $1
	`
	
	rows, err := s.db.QueryContext(ctx, query, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()
	
	// Check each key with matching prefix
	for rows.Next() {
		key := &models.APIKey{}
		err := rows.Scan(
			&key.ID, &key.TenantID, &key.Name, &key.Prefix, &key.Hash,
			&key.LastUsedAt, &key.CreatedAt, &key.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		// Validate the full key against hash
		if bcrypt.CompareHashAndPassword([]byte(key.Hash), []byte(keyValue)) == nil {
			// Update last used timestamp
			s.updateLastUsed(ctx, key.ID)
			return key, nil
		}
	}
	
	return nil, fmt.Errorf("invalid API key")
}

// Delete deletes an API key
func (s *Service) Delete(ctx context.Context, id int, tenantID int) error {
	_, err := s.db.ExecContext(ctx, 
		"DELETE FROM api_keys WHERE id = $1 AND tenant_id = $2", 
		id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}
	
	return nil
}

// generateAPIKey generates a secure random API key
func (s *Service) generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	key := base64.URLEncoding.EncodeToString(bytes)
	// Remove padding and make URL-safe
	key = strings.TrimRight(key, "=")
	return "el_" + key, nil
}

// updateLastUsed updates the last used timestamp for an API key
func (s *Service) updateLastUsed(ctx context.Context, id int) {
	_, err := s.db.ExecContext(ctx, 
		"UPDATE api_keys SET last_used_at = NOW() WHERE id = $1", id)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update last_used_at for API key %d: %v\n", id, err)
	}
}