// Package customdomains provides custom domain management functionality for the edge.link proxy service.
package customdomains

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
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
	defer func() { _ = rows.Close() }()

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
//
//nolint:unused // used for potential future extensions and kept for completeness
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
//
//nolint:unused // not used in current flow; kept for future DNS verification feature
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

// UpdateStatus updates the status of a custom domain
func (s *Service) UpdateStatus(ctx context.Context, id int, tenantID int, status string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE custom_domains SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND tenant_id = $3",
		status, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update custom domain status: %w", err)
	}

	return nil
}

// VerifyDomain verifies domain ownership via DNS TXT record or HTTP challenge
func (s *Service) VerifyDomain(ctx context.Context, id int, tenantID int) (bool, error) {
	// Get domain record
	domains, err := s.GetByTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	var domain *models.CustomDomain
	for _, d := range domains {
		if d.ID == id {
			domain = d
			break
		}
	}

	if domain == nil {
		return false, fmt.Errorf("domain not found")
	}

	// Check DNS TXT record first
	verified, err := s.checkDNSVerification(domain.Hostname, domain.VerificationToken)
	if err == nil && verified {
		// Update status to verified
		if err := s.UpdateStatus(ctx, id, tenantID, "verified"); err != nil {
			return false, fmt.Errorf("failed to update domain status: %w", err)
		}
		return true, nil
	}

	// If DNS verification fails, try HTTP challenge
	verified = s.checkHTTPVerification(domain.Hostname, domain.VerificationToken)
	if verified {
		// Update status to verified
		if err := s.UpdateStatus(ctx, id, tenantID, "verified"); err != nil {
			return false, fmt.Errorf("failed to update domain status: %w", err)
		}
		return true, nil
	}

	// Update status to failed
	if err := s.UpdateStatus(ctx, id, tenantID, "failed"); err != nil {
		return false, fmt.Errorf("failed to update domain status: %w", err)
	}

	return false, nil
}

// checkHTTPVerification checks if domain has the verification token via HTTP challenge
func (s *Service) checkHTTPVerification(hostname, token string) bool {
	// Try both HTTP and HTTPS
	for _, scheme := range []string{"http", "https"} {
		url := fmt.Sprintf("%s://%s/.well-known/edge-link.txt", scheme, hostname)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println("Error closing response body:", err)
			}
		}(resp.Body)

		if resp.StatusCode == http.StatusOK {
			// Read response body and check if it contains our token
			body := make([]byte, 1024)
			n, _ := resp.Body.Read(body)
			content := string(body[:n])
			if strings.Contains(content, token) {
				return true
			}
		}
	}
	return false
}
