// Package customdomains provides custom domain management functionality for the edge.link proxy service.
package customdomains

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/ubcent/edge.link/internal/models"
)

// Constants for token generation
const (
	verificationTokenBytes = 24
)

// Service provides custom domain management functionality
type Service struct {
	db *sql.DB
}

// NewService creates a new custom domain service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Create creates a new custom domain
func (s *Service) Create(ctx context.Context, domain *models.CustomDomain) error {
	// Generate verification token
	token, err := s.generateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}
	domain.VerificationToken = token

	query := `
		INSERT INTO custom_domains (tenant_id, hostname, verification_token, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err = s.db.QueryRowContext(ctx, query,
		domain.TenantID, domain.Hostname, domain.VerificationToken, domain.Status,
	).Scan(&domain.ID, &domain.CreatedAt, &domain.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create custom domain: %w", err)
	}

	return nil
}

// GetByTenant retrieves custom domains for a specific tenant
func (s *Service) GetByTenant(ctx context.Context, tenantID int) ([]*models.CustomDomain, error) {
	query := `
		SELECT id, tenant_id, hostname, verification_token, status, created_at, updated_at
		FROM custom_domains
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom domains for tenant: %w", err)
	}
	defer rows.Close()

	var domains []*models.CustomDomain
	for rows.Next() {
		domain := &models.CustomDomain{}
		err := rows.Scan(
			&domain.ID, &domain.TenantID, &domain.Hostname, &domain.VerificationToken,
			&domain.Status, &domain.CreatedAt, &domain.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan custom domain: %w", err)
		}
		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over custom domain rows: %w", err)
	}

	return domains, nil
}

// GetByHostname retrieves a custom domain by hostname
func (s *Service) GetByHostname(ctx context.Context, hostname string) (*models.CustomDomain, error) {
	query := `
		SELECT id, tenant_id, hostname, verification_token, status, created_at, updated_at
		FROM custom_domains
		WHERE hostname = $1
	`
	return s.queryCustomDomain(ctx, query, hostname)
}

// getByID retrieves a custom domain by ID
func (s *Service) getByID(ctx context.Context, id int) (*models.CustomDomain, error) {
	query := `
		SELECT id, tenant_id, hostname, verification_token, status, created_at, updated_at
		FROM custom_domains
		WHERE id = $1
	`
	return s.queryCustomDomain(ctx, query, id)
}

// queryCustomDomain is a helper function to query custom domain by different criteria
func (s *Service) queryCustomDomain(ctx context.Context, query string, args ...interface{}) (*models.CustomDomain, error) {
	domain := &models.CustomDomain{}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&domain.ID, &domain.TenantID, &domain.Hostname, &domain.VerificationToken,
		&domain.Status, &domain.CreatedAt, &domain.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("custom domain not found")
		}
		return nil, fmt.Errorf("failed to get custom domain: %w", err)
	}

	return domain, nil
}

// generateVerificationToken generates a random verification token
func (s *Service) generateVerificationToken() (string, error) {
	bytes := make([]byte, verificationTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// checkDNSVerification checks if domain has the required TXT record
func (s *Service) checkDNSVerification(domain, token string) (bool, error) {
	// Look for TXT record at _edgelink.domain.com
	verifyDomain := fmt.Sprintf("_edgelink.%s", domain)

	txtRecords, err := net.LookupTXT(verifyDomain)
	if err != nil {
		return false, fmt.Errorf("failed to lookup TXT records: %w", err)
	}

	expectedValue := fmt.Sprintf("edgelink-verify=%s", token)
	for _, record := range txtRecords {
		if strings.Contains(record, expectedValue) {
			return true, nil
		}
	}

	return false, nil
}

// Delete deletes a custom domain
func (s *Service) Delete(ctx context.Context, id int, tenantID int) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM custom_domains WHERE id = $1 AND tenant_id = $2",
		id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete custom domain: %w", err)
	}

	return nil
}
