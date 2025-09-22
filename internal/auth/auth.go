package auth

import (
	"crypto/subtle"
	"errors"
	"strings"
	"sync"
)

var (
	// ErrInvalidAPIKey is returned when an API key is invalid
	ErrInvalidAPIKey = errors.New("invalid API key")
	// ErrPermissionDenied is returned when access is denied
	ErrPermissionDenied = errors.New("permission denied")
	// ErrAPIKeyDisabled is returned when an API key is disabled
	ErrAPIKeyDisabled = errors.New("API key is disabled")
)

// APIKey represents an API key with permissions
type APIKey struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	RateLimit   int      `json:"rate_limit"`
	Enabled     bool     `json:"enabled"`
}

// HasPermission checks if the API key has a specific permission
func (k *APIKey) HasPermission(permission string) bool {
	if !k.Enabled {
		return false
	}

	for _, p := range k.Permissions {
		if p == "*" || p == permission {
			return true
		}
		// Support wildcard permissions like "proxy.*"
		if strings.HasSuffix(p, "*") {
			prefix := strings.TrimSuffix(p, "*")
			if strings.HasPrefix(permission, prefix) {
				return true
			}
		}
	}
	return false
}

// Manager manages API keys and authentication
type Manager struct {
	mu   sync.RWMutex
	keys map[string]*APIKey
}

// NewManager creates a new authentication manager
func NewManager() *Manager {
	return &Manager{
		keys: make(map[string]*APIKey),
	}
}

// AddKey adds an API key to the manager
func (m *Manager) AddKey(key *APIKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[key.Key] = key
}

// RemoveKey removes an API key from the manager
func (m *Manager) RemoveKey(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.keys, key)
}

// GetKey retrieves an API key by its value
func (m *Manager) GetKey(key string) (*APIKey, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	apiKey, exists := m.keys[key]
	if !exists {
		return nil, false
	}
	
	// Return a copy to prevent external modification
	keyCopy := *apiKey
	return &keyCopy, true
}

// ValidateKey validates an API key and returns the associated APIKey
func (m *Manager) ValidateKey(key string) (*APIKey, error) {
	if key == "" {
		return nil, ErrInvalidAPIKey
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Use constant-time comparison to prevent timing attacks
	for storedKey, apiKey := range m.keys {
		if subtle.ConstantTimeCompare([]byte(key), []byte(storedKey)) == 1 {
			if !apiKey.Enabled {
				return nil, ErrAPIKeyDisabled
			}
			
			// Return a copy
			keyCopy := *apiKey
			return &keyCopy, nil
		}
	}

	return nil, ErrInvalidAPIKey
}

// CheckPermission validates an API key and checks if it has the required permission
func (m *Manager) CheckPermission(key, permission string) (*APIKey, error) {
	apiKey, err := m.ValidateKey(key)
	if err != nil {
		return nil, err
	}

	if !apiKey.HasPermission(permission) {
		return nil, ErrPermissionDenied
	}

	return apiKey, nil
}

// ListKeys returns all API keys (without the actual key values for security)
func (m *Manager) ListKeys() []KeyInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]KeyInfo, 0, len(m.keys))
	for _, apiKey := range m.keys {
		keys = append(keys, KeyInfo{
			Name:        apiKey.Name,
			Permissions: apiKey.Permissions,
			RateLimit:   apiKey.RateLimit,
			Enabled:     apiKey.Enabled,
		})
	}

	return keys
}

// KeyInfo represents API key information without the actual key
type KeyInfo struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	RateLimit   int      `json:"rate_limit"`
	Enabled     bool     `json:"enabled"`
}

// UpdateKeyPermissions updates the permissions for an API key
func (m *Manager) UpdateKeyPermissions(key string, permissions []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	apiKey, exists := m.keys[key]
	if !exists {
		return ErrInvalidAPIKey
	}

	apiKey.Permissions = permissions
	return nil
}

// EnableKey enables or disables an API key
func (m *Manager) EnableKey(key string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	apiKey, exists := m.keys[key]
	if !exists {
		return ErrInvalidAPIKey
	}

	apiKey.Enabled = enabled
	return nil
}

// Stats returns authentication statistics
func (m *Manager) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enabled := 0
	disabled := 0

	for _, key := range m.keys {
		if key.Enabled {
			enabled++
		} else {
			disabled++
		}
	}

	return Stats{
		TotalKeys:    len(m.keys),
		EnabledKeys:  enabled,
		DisabledKeys: disabled,
	}
}

// Stats represents authentication statistics
type Stats struct {
	TotalKeys    int `json:"total_keys"`
	EnabledKeys  int `json:"enabled_keys"`
	DisabledKeys int `json:"disabled_keys"`
}

// ExtractAPIKey extracts the API key from various sources in the request
func ExtractAPIKey(authHeader, queryParam, headerKey string) string {
	// Try Authorization header first (Bearer token)
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		if strings.HasPrefix(authHeader, "APIKey ") {
			return strings.TrimPrefix(authHeader, "APIKey ")
		}
		// Fallback to raw header value
		return authHeader
	}

	// Try query parameter
	if queryParam != "" {
		return queryParam
	}

	// Try custom header
	if headerKey != "" {
		return headerKey
	}

	return ""
}