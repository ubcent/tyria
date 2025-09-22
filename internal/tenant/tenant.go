package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Domain    string    `json:"domain" db:"domain"`
	Plan      string    `json:"plan" db:"plan"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Service provides tenant management functionality
type Service struct {
	db *sql.DB
}

// NewService creates a new tenant service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Create creates a new tenant
func (s *Service) Create(ctx context.Context, tenant *Tenant) error {
	query := `
		INSERT INTO tenants (name, email, domain, plan, enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	
	err := s.db.QueryRowContext(ctx, query,
		tenant.Name, tenant.Email, tenant.Domain, tenant.Plan, tenant.Enabled,
	).Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	
	return nil
}

// GetByID retrieves a tenant by ID
func (s *Service) GetByID(ctx context.Context, id int) (*Tenant, error) {
	query := `
		SELECT id, name, email, domain, plan, enabled, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`
	
	tenant := &Tenant{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.Email, &tenant.Domain,
		&tenant.Plan, &tenant.Enabled, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	return tenant, nil
}

// GetByDomain retrieves a tenant by domain
func (s *Service) GetByDomain(ctx context.Context, domain string) (*Tenant, error) {
	query := `
		SELECT id, name, email, domain, plan, enabled, created_at, updated_at
		FROM tenants
		WHERE domain = $1 AND enabled = true
	`
	
	tenant := &Tenant{}
	err := s.db.QueryRowContext(ctx, query, domain).Scan(
		&tenant.ID, &tenant.Name, &tenant.Email, &tenant.Domain,
		&tenant.Plan, &tenant.Enabled, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tenant not found for domain: %s", domain)
		}
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}
	
	return tenant, nil
}

// List retrieves all tenants with pagination
func (s *Service) List(ctx context.Context, limit, offset int) ([]*Tenant, error) {
	query := `
		SELECT id, name, email, domain, plan, enabled, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()
	
	var tenants []*Tenant
	for rows.Next() {
		tenant := &Tenant{}
		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Email, &tenant.Domain,
			&tenant.Plan, &tenant.Enabled, &tenant.CreatedAt, &tenant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

// Update updates a tenant
func (s *Service) Update(ctx context.Context, tenant *Tenant) error {
	query := `
		UPDATE tenants 
		SET name = $1, email = $2, domain = $3, plan = $4, enabled = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`
	
	err := s.db.QueryRowContext(ctx, query,
		tenant.Name, tenant.Email, tenant.Domain, tenant.Plan, tenant.Enabled, tenant.ID,
	).Scan(&tenant.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	
	return nil
}

// Delete deletes a tenant
func (s *Service) Delete(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM tenants WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	
	return nil
}