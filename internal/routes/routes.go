// Package routes provides route configuration and management functionality for the edge.link proxy service.
package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ubcent/edge.link/internal/models"
)

// Service provides route management functionality
type Service struct {
	db *sql.DB
}

// NewService creates a new route service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Create creates a new route
func (s *Service) Create(ctx context.Context, route *models.Route) error {
	query := `
		INSERT INTO routes (tenant_id, name, match_path, upstream_url, headers_json, auth_mode, caching_policy_json, rate_limit_policy_json, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		route.TenantID, route.Name, route.MatchPath, route.UpstreamURL,
		route.HeadersJSON, route.AuthMode, route.CachingPolicyJSON,
		route.RateLimitPolicyJSON, route.Enabled,
	).Scan(&route.ID, &route.CreatedAt, &route.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create route: %w", err)
	}

	return nil
}

// GetByID retrieves a route by ID
func (s *Service) GetByID(ctx context.Context, id int) (*models.Route, error) {
	query := `
		SELECT id, tenant_id, name, match_path, upstream_url, headers_json, auth_mode, 
		       caching_policy_json, rate_limit_policy_json, enabled, created_at, updated_at
		FROM routes
		WHERE id = $1
	`

	route := &models.Route{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&route.ID, &route.TenantID, &route.Name, &route.MatchPath, &route.UpstreamURL,
		&route.HeadersJSON, &route.AuthMode, &route.CachingPolicyJSON,
		&route.RateLimitPolicyJSON, &route.Enabled, &route.CreatedAt, &route.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return route, nil
}

// GetByTenant retrieves routes for a specific tenant
func (s *Service) GetByTenant(ctx context.Context, tenantID int) ([]*models.Route, error) {
	query := `
		SELECT id, tenant_id, name, match_path, upstream_url, headers_json, auth_mode,
		       caching_policy_json, rate_limit_policy_json, enabled, created_at, updated_at
		FROM routes
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes for tenant: %w", err)
	}
	defer rows.Close()

	var routes []*models.Route
	for rows.Next() {
		route := &models.Route{}
		err := rows.Scan(
			&route.ID, &route.TenantID, &route.Name, &route.MatchPath, &route.UpstreamURL,
			&route.HeadersJSON, &route.AuthMode, &route.CachingPolicyJSON,
			&route.RateLimitPolicyJSON, &route.Enabled, &route.CreatedAt, &route.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over routes: %w", err)
	}

	return routes, nil
}

// GetByPath retrieves a route by tenant and path
func (s *Service) GetByPath(ctx context.Context, tenantID int, path string) (*models.Route, error) {
	query := `
		SELECT id, tenant_id, name, match_path, upstream_url, headers_json, auth_mode,
		       caching_policy_json, rate_limit_policy_json, enabled, created_at, updated_at
		FROM routes
		WHERE tenant_id = $1 AND match_path = $2 AND enabled = true
	`

	route := &models.Route{}
	err := s.db.QueryRowContext(ctx, query, tenantID, path).Scan(
		&route.ID, &route.TenantID, &route.Name, &route.MatchPath, &route.UpstreamURL,
		&route.HeadersJSON, &route.AuthMode, &route.CachingPolicyJSON,
		&route.RateLimitPolicyJSON, &route.Enabled, &route.CreatedAt, &route.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("route not found")
		}
		return nil, fmt.Errorf("failed to get route by path: %w", err)
	}

	return route, nil
}

// Update updates an existing route
func (s *Service) Update(ctx context.Context, route *models.Route) error {
	query := `
		UPDATE routes 
		SET name = $1, match_path = $2, upstream_url = $3, headers_json = $4, auth_mode = $5,
		    caching_policy_json = $6, rate_limit_policy_json = $7, enabled = $8, updated_at = NOW()
		WHERE id = $9 AND tenant_id = $10
		RETURNING updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		route.Name, route.MatchPath, route.UpstreamURL, route.HeadersJSON,
		route.AuthMode, route.CachingPolicyJSON, route.RateLimitPolicyJSON,
		route.Enabled, route.ID, route.TenantID,
	).Scan(&route.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}

	return nil
}

// Delete deletes a route
func (s *Service) Delete(ctx context.Context, id int, tenantID int) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM routes WHERE id = $1 AND tenant_id = $2",
		id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}

	return nil
}

// CreateDefaultCachingPolicy creates a default caching policy JSON
func CreateDefaultCachingPolicy(enabled bool, ttlSeconds int) (json.RawMessage, error) {
	policy := models.CachingPolicy{
		Enabled:    enabled,
		TTLSeconds: ttlSeconds,
	}
	return json.Marshal(policy)
}

// CreateDefaultRateLimitPolicy creates a default rate limit policy JSON
func CreateDefaultRateLimitPolicy(enabled bool, requestsPerMinute int) (json.RawMessage, error) {
	policy := models.RateLimitPolicy{
		Enabled:           enabled,
		RequestsPerMinute: requestsPerMinute,
	}
	return json.Marshal(policy)
}

// CreateDefaultHeaders creates default headers JSON
func CreateDefaultHeaders(headers map[string]string) (json.RawMessage, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	return json.Marshal(headers)
}
