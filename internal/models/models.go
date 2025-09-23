package models

import (
	"encoding/json"
	"time"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Plan      string    `json:"plan" db:"plan"`     // free, pro, enterprise
	Status    string    `json:"status" db:"status"` // active, suspended, canceled
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID             int       `json:"id" db:"id"`
	TenantID       int       `json:"tenant_id" db:"tenant_id"`
	Email          string    `json:"email" db:"email"`
	HashedPassword string    `json:"-" db:"hashed_password"` // Hidden from JSON
	Role           string    `json:"role" db:"role"`         // owner, admin, viewer
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// APIKey represents an API key for tenant authentication
type APIKey struct {
	ID         int        `json:"id" db:"id"`
	TenantID   int        `json:"tenant_id" db:"tenant_id"`
	Name       string     `json:"name" db:"name"`
	Prefix     string     `json:"prefix" db:"prefix"` // Visible part of key
	Hash       string     `json:"-" db:"hash"`        // Hidden hashed key
	LastUsedAt *time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// Route represents a proxy route configuration
type Route struct {
	ID                  int             `json:"id" db:"id"`
	TenantID            int             `json:"tenant_id" db:"tenant_id"`
	Name                string          `json:"name" db:"name"`
	MatchPath           string          `json:"match_path" db:"match_path"`
	UpstreamURL         string          `json:"upstream_url" db:"upstream_url"`
	HeadersJSON         json.RawMessage `json:"headers_json" db:"headers_json"`
	AuthMode            string          `json:"auth_mode" db:"auth_mode"` // none, api_key, bearer
	CachingPolicyJSON   json.RawMessage `json:"caching_policy_json" db:"caching_policy_json"`
	RateLimitPolicyJSON json.RawMessage `json:"rate_limit_policy_json" db:"rate_limit_policy_json"`
	Enabled             bool            `json:"enabled" db:"enabled"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// CachingPolicy represents the caching configuration for a route
type CachingPolicy struct {
	Enabled     bool     `json:"enabled"`
	TTLSeconds  int      `json:"ttl_seconds"`
	VaryHeaders []string `json:"vary_headers"`
}

// RateLimitPolicy represents the rate limiting configuration for a route
type RateLimitPolicy struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
}

// GetCachingPolicy parses the caching policy JSON
func (r *Route) GetCachingPolicy() (*CachingPolicy, error) {
	var policy CachingPolicy
	if err := json.Unmarshal(r.CachingPolicyJSON, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// GetRateLimitPolicy parses the rate limit policy JSON
func (r *Route) GetRateLimitPolicy() (*RateLimitPolicy, error) {
	var policy RateLimitPolicy
	if err := json.Unmarshal(r.RateLimitPolicyJSON, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// GetHeaders parses the headers JSON
func (r *Route) GetHeaders() (map[string]string, error) {
	var headers map[string]string
	if err := json.Unmarshal(r.HeadersJSON, &headers); err != nil {
		return nil, err
	}
	return headers, nil
}

// CustomDomain represents a custom domain for a tenant
type CustomDomain struct {
	ID                int       `json:"id" db:"id"`
	TenantID          int       `json:"tenant_id" db:"tenant_id"`
	Hostname          string    `json:"hostname" db:"hostname"`
	VerificationToken string    `json:"verification_token" db:"verification_token"`
	Status            string    `json:"status" db:"status"` // pending, verified, failed
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// RequestLog represents a logged request for analytics
type RequestLog struct {
	ID          int       `json:"id" db:"id"`
	TenantID    int       `json:"tenant_id" db:"tenant_id"`
	RouteID     *int      `json:"route_id" db:"route_id"`
	StatusCode  int       `json:"status_code" db:"status_code"`
	LatencyMs   int       `json:"latency_ms" db:"latency_ms"`
	CacheStatus string    `json:"cache_status" db:"cache_status"` // hit, miss, bypass
	BytesIn     int       `json:"bytes_in" db:"bytes_in"`
	BytesOut    int       `json:"bytes_out" db:"bytes_out"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
