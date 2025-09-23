package domains

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"
)

// Domain represents a domain configuration for a tenant
type Domain struct {
	ID          int       `json:"id" db:"id"`
	TenantID    int       `json:"tenant_id" db:"tenant_id"`
	Domain      string    `json:"domain" db:"domain"`
	ProxyURL    string    `json:"proxy_url" db:"proxy_url"`
	Verified    bool      `json:"verified" db:"verified"`
	VerifyToken string    `json:"verify_token" db:"verify_token"`
	SSLEnabled  bool      `json:"ssl_enabled" db:"ssl_enabled"`
	SSLCertPath string    `json:"ssl_cert_path" db:"ssl_cert_path"`
	SSLKeyPath  string    `json:"ssl_key_path" db:"ssl_key_path"`
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Service provides domain management functionality
type Service struct {
	db         *sql.DB
	baseDomain string
}

// NewService creates a new domain service
func NewService(db *sql.DB, baseDomain string) *Service {
	return &Service{
		db:         db,
		baseDomain: baseDomain,
	}
}

// Create creates a new domain configuration
func (s *Service) Create(ctx context.Context, domain *Domain) error {
	// Generate verification token
	token, err := generateVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}
	domain.VerifyToken = token

	// Generate proxy URL based on domain
	domain.ProxyURL = s.generateProxyURL(domain.Domain)

	query := `
		INSERT INTO domains (tenant_id, domain, proxy_url, verify_token, verified, ssl_enabled, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err = s.db.QueryRowContext(ctx, query,
		domain.TenantID, domain.Domain, domain.ProxyURL, domain.VerifyToken,
		domain.Verified, domain.SSLEnabled, domain.Enabled,
	).Scan(&domain.ID, &domain.CreatedAt, &domain.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}

	return nil
}

// GetByTenant retrieves domains for a specific tenant
func (s *Service) GetByTenant(ctx context.Context, tenantID int) ([]*Domain, error) {
	query := `
		SELECT id, tenant_id, domain, proxy_url, verified, verify_token,
		       ssl_enabled, ssl_cert_path, ssl_key_path, enabled, created_at, updated_at
		FROM domains
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domains for tenant: %w", err)
	}
	defer rows.Close()

	var domains []*Domain
	for rows.Next() {
		domain := &Domain{}
		err := rows.Scan(
			&domain.ID, &domain.TenantID, &domain.Domain, &domain.ProxyURL,
			&domain.Verified, &domain.VerifyToken, &domain.SSLEnabled,
			&domain.SSLCertPath, &domain.SSLKeyPath, &domain.Enabled,
			&domain.CreatedAt, &domain.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

// VerifyDomain verifies domain ownership via DNS TXT record
func (s *Service) VerifyDomain(ctx context.Context, domainID int) error {
	// Get domain record
	domain, err := s.getByID(ctx, domainID)
	if err != nil {
		return err
	}

	// Check DNS TXT record for verification token
	verified, err := s.checkDNSVerification(domain.Domain, domain.VerifyToken)
	if err != nil {
		return fmt.Errorf("failed to verify domain: %w", err)
	}

	if verified {
		// Update domain as verified
		query := `UPDATE domains SET verified = true, updated_at = NOW() WHERE id = $1`
		_, err = s.db.ExecContext(ctx, query, domainID)
		if err != nil {
			return fmt.Errorf("failed to update domain verification: %w", err)
		}
	}

	return nil
}

// getByID retrieves a domain by ID
func (s *Service) getByID(ctx context.Context, id int) (*Domain, error) {
	query := `
		SELECT id, tenant_id, domain, proxy_url, verified, verify_token,
		       ssl_enabled, ssl_cert_path, ssl_key_path, enabled, created_at, updated_at
		FROM domains
		WHERE id = $1
	`

	domain := &Domain{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&domain.ID, &domain.TenantID, &domain.Domain, &domain.ProxyURL,
		&domain.Verified, &domain.VerifyToken, &domain.SSLEnabled,
		&domain.SSLCertPath, &domain.SSLKeyPath, &domain.Enabled,
		&domain.CreatedAt, &domain.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("domain not found")
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return domain, nil
}

// generateProxyURL generates a proxy URL from a domain
func (s *Service) generateProxyURL(domain string) string {
	// Convert domain to proxy subdomain format
	// e.g., api.example.com -> api-example-com.edge.link
	subdomain := strings.ReplaceAll(domain, ".", "-")
	return fmt.Sprintf("https://%s.%s", subdomain, s.baseDomain)
}

// generateVerificationToken generates a random verification token
func generateVerificationToken() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 32

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
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
		if record == expectedValue {
			return true, nil
		}
	}

	return false, nil
}
