// Package admin provides administrative server functionality for the edge.link proxy service.
// It includes authentication, user management, route configuration, and API key management.
package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/ubcent/edge.link/internal/apikeys"
	"github.com/ubcent/edge.link/internal/models"
	"github.com/ubcent/edge.link/internal/routes"
)

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// writeJSON is a helper function to encode JSON responses with proper error handling
func writeJSON(w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

// Server represents the admin API server
type Server struct {
	db             *sql.DB
	router         *mux.Router
	apiKeysService *apikeys.Service
	routesService  *routes.Service
}

// writeJSONError writes a JSON formatted error response
func (s *Server) writeJSONError(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	errorResp := ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	}
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ProxyRoute represents a proxy route configuration
type ProxyRoute struct {
	ID                       int       `json:"id"`
	Path                     string    `json:"path"`
	Target                   string    `json:"target"`
	Methods                  []string  `json:"methods"`
	CacheEnabled             bool      `json:"cache_enabled"`
	CacheTTL                 int       `json:"cache_ttl"`
	RateLimitEnabled         bool      `json:"rate_limit_enabled"`
	RateLimitRate            int       `json:"rate_limit_rate"`
	RateLimitBurst           int       `json:"rate_limit_burst"`
	RateLimitPeriod          int       `json:"rate_limit_period"`
	RateLimitPerClient       bool      `json:"rate_limit_per_client"`
	AuthRequired             bool      `json:"auth_required"`
	AuthKeys                 []string  `json:"auth_keys"`
	ValidationEnabled        bool      `json:"validation_enabled"`
	ValidationRequestSchema  *string   `json:"validation_request_schema"`
	ValidationResponseSchema *string   `json:"validation_response_schema"`
	Enabled                  bool      `json:"enabled"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// APIKey represents an API key configuration
type APIKey struct {
	ID          int        `json:"id"`
	KeyValue    string     `json:"key_value"`
	Name        string     `json:"name"`
	Permissions []string   `json:"permissions"`
	RateLimit   int        `json:"rate_limit"`
	Enabled     bool       `json:"enabled"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"` // The full key is only returned once
	Prefix    string    `json:"prefix"`
	CreatedAt time.Time `json:"created_at"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalRequests   int64   `json:"total_requests"`
	AvgResponseTime float64 `json:"avg_response_time"`
	SuccessRate     float64 `json:"success_rate"`
	CacheHitRate    float64 `json:"cache_hit_rate"`
	ActiveRoutes    int     `json:"active_routes"`
	ActiveAPIKeys   int     `json:"active_api_keys"`
}

// NewServer creates a new admin API server
func NewServer(databaseURL string) (*Server, error) {
	var driverName string
	if strings.Contains(databaseURL, "file:") || strings.Contains(databaseURL, "sqlite") {
		driverName = "sqlite"
	} else {
		driverName = "postgres"
	}

	db, err := sql.Open(driverName, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	server := &Server{
		db:             db,
		router:         mux.NewRouter(),
		apiKeysService: apikeys.NewService(db),
		routesService:  routes.NewService(db),
	}

	server.setupRoutes()
	return server, nil
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api").Subrouter()

	// CORS middleware
	api.Use(s.corsMiddleware)

	// Auth routes
	authServer := NewAuthServer(s.db)
	api.HandleFunc("/auth/signup", authServer.HandleSignup).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/signin", authServer.HandleSignin).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/signout", authServer.HandleSignout).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/confirm-email", authServer.HandleConfirmEmail).Methods("POST", "OPTIONS")

	// Protected auth routes
	authMiddleware := authServer.GetAuthMiddleware()
	protectedAuth := api.PathPrefix("/auth").Subrouter()
	protectedAuth.Use(authMiddleware.RequireAuth)
	protectedAuth.HandleFunc("/profile", authServer.HandleProfile).Methods("GET", "OPTIONS")

	// Routes management
	api.HandleFunc("/routes", s.handleRoutes).Methods("GET", "POST", "OPTIONS")
	api.HandleFunc("/routes/{id}", s.handleRoute).Methods("GET", "PUT", "DELETE", "OPTIONS")

	// API Keys management - v1 endpoints
	v1 := api.PathPrefix("/v1").Subrouter()
	v1.HandleFunc("/api-keys", s.handleAPIKeys).Methods("GET", "POST", "OPTIONS")
	v1.HandleFunc("/api-keys/{id}", s.handleAPIKey).Methods("DELETE", "OPTIONS")

	// Routes management - v1 endpoints with validation
	v1.HandleFunc("/routes", s.handleV1Routes).Methods("GET", "POST", "OPTIONS")
	v1.HandleFunc("/routes/{id}", s.handleV1Route).Methods("GET", "PUT", "DELETE", "OPTIONS")

	// Legacy API Keys management (for backward compatibility)
	api.HandleFunc("/keys", s.handleAPIKeys).Methods("GET", "POST", "OPTIONS")
	api.HandleFunc("/keys/{id}", s.handleAPIKey).Methods("GET", "PUT", "DELETE", "OPTIONS")

	// Dashboard
	api.HandleFunc("/dashboard/stats", s.handleDashboardStats).Methods("GET", "OPTIONS")
	api.HandleFunc("/dashboard/activity", s.handleDashboardActivity).Methods("GET", "OPTIONS")

	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

// getTenantID extracts tenant ID from request context
// For now, we'll use a hardcoded tenant ID (1) since auth is not fully implemented
// In a real implementation, this would come from the authenticated user context
func (s *Server) getTenantID(r *http.Request) int {
	// TODO: Extract from authenticated user context
	return 1
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleRoutes handles routes collection endpoints
func (s *Server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getRoutes(w, r)
	case "POST":
		s.createRoute(w, r)
	}
}

// handleRoute handles individual route endpoints
func (s *Server) handleRoute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeJSONError(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.getRoute(w, r, id)
	case "PUT":
		s.updateRoute(w, r, id)
	case "DELETE":
		s.deleteRoute(w, r, id)
	}
}

// getRoutes retrieves all routes
func (s *Server) getRoutes(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, path, target, methods, cache_enabled, cache_ttl,
		       rate_limit_enabled, rate_limit_rate, rate_limit_burst, rate_limit_period,
		       rate_limit_per_client, auth_required, auth_keys, validation_enabled,
		       validation_request_schema, validation_response_schema, enabled,
		       created_at, updated_at
		FROM proxy_routes
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("Error querying routes: %v", err)
		s.writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var routes []ProxyRoute
	for rows.Next() {
		var route ProxyRoute
		var methods, authKeys pq.StringArray

		err := rows.Scan(
			&route.ID, &route.Path, &route.Target, &methods, &route.CacheEnabled, &route.CacheTTL,
			&route.RateLimitEnabled, &route.RateLimitRate, &route.RateLimitBurst, &route.RateLimitPeriod,
			&route.RateLimitPerClient, &route.AuthRequired, &authKeys, &route.ValidationEnabled,
			&route.ValidationRequestSchema, &route.ValidationResponseSchema, &route.Enabled,
			&route.CreatedAt, &route.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning route: %v", err)
			continue
		}

		// Convert pq.StringArray to []string
		route.Methods = []string(methods)
		route.AuthKeys = []string(authKeys)

		routes = append(routes, route)
	}

	writeJSON(w, routes)
}

// createRoute creates a new route
func (s *Server) createRoute(w http.ResponseWriter, r *http.Request) {
	var route ProxyRoute
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		s.writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert to PostgreSQL arrays
	methodsArray := pq.StringArray(route.Methods)
	authKeysArray := pq.StringArray(route.AuthKeys)

	query := `
		INSERT INTO proxy_routes (
			path, target, methods, cache_enabled, cache_ttl,
			rate_limit_enabled, rate_limit_rate, rate_limit_burst, rate_limit_period,
			rate_limit_per_client, auth_required, auth_keys, validation_enabled,
			validation_request_schema, validation_response_schema, enabled
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRow(query,
		route.Path, route.Target, methodsArray, route.CacheEnabled, route.CacheTTL,
		route.RateLimitEnabled, route.RateLimitRate, route.RateLimitBurst, route.RateLimitPeriod,
		route.RateLimitPerClient, route.AuthRequired, authKeysArray, route.ValidationEnabled,
		route.ValidationRequestSchema, route.ValidationResponseSchema, route.Enabled,
	).Scan(&route.ID, &route.CreatedAt, &route.UpdatedAt)

	if err != nil {
		log.Printf("Error creating route: %v", err)
		s.writeJSONError(w, "Failed to create route", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, route)
}

// updateRoute updates an existing route
func (s *Server) updateRoute(w http.ResponseWriter, r *http.Request, id int) {
	var route ProxyRoute
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		s.writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert to PostgreSQL arrays
	methodsArray := pq.StringArray(route.Methods)
	authKeysArray := pq.StringArray(route.AuthKeys)

	query := `
		UPDATE proxy_routes SET
			path = $1, target = $2, methods = $3, cache_enabled = $4, cache_ttl = $5,
			rate_limit_enabled = $6, rate_limit_rate = $7, rate_limit_burst = $8, rate_limit_period = $9,
			rate_limit_per_client = $10, auth_required = $11, auth_keys = $12, validation_enabled = $13,
			validation_request_schema = $14, validation_response_schema = $15, enabled = $16,
			updated_at = NOW()
		WHERE id = $17
		RETURNING updated_at
	`

	err := s.db.QueryRow(query,
		route.Path, route.Target, methodsArray, route.CacheEnabled, route.CacheTTL,
		route.RateLimitEnabled, route.RateLimitRate, route.RateLimitBurst, route.RateLimitPeriod,
		route.RateLimitPerClient, route.AuthRequired, authKeysArray, route.ValidationEnabled,
		route.ValidationRequestSchema, route.ValidationResponseSchema, route.Enabled, id,
	).Scan(&route.UpdatedAt)

	if err != nil {
		log.Printf("Error updating route: %v", err)
		s.writeJSONError(w, "Failed to update route", http.StatusInternalServerError)
		return
	}

	route.ID = id
	writeJSON(w, route)
}

// deleteRoute deletes a route
func (s *Server) deleteRoute(w http.ResponseWriter, r *http.Request, id int) {
	_, err := s.db.Exec("DELETE FROM proxy_routes WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting route: %v", err)
		s.writeJSONError(w, "Failed to delete route", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getRoute retrieves a single route
func (s *Server) getRoute(w http.ResponseWriter, r *http.Request, id int) {
	query := `
		SELECT id, path, target, methods, cache_enabled, cache_ttl,
		       rate_limit_enabled, rate_limit_rate, rate_limit_burst, rate_limit_period,
		       rate_limit_per_client, auth_required, auth_keys, validation_enabled,
		       validation_request_schema, validation_response_schema, enabled,
		       created_at, updated_at
		FROM proxy_routes
		WHERE id = $1
	`

	var route ProxyRoute
	var methods, authKeys pq.StringArray

	err := s.db.QueryRow(query, id).Scan(
		&route.ID, &route.Path, &route.Target, &methods, &route.CacheEnabled, &route.CacheTTL,
		&route.RateLimitEnabled, &route.RateLimitRate, &route.RateLimitBurst, &route.RateLimitPeriod,
		&route.RateLimitPerClient, &route.AuthRequired, &authKeys, &route.ValidationEnabled,
		&route.ValidationRequestSchema, &route.ValidationResponseSchema, &route.Enabled,
		&route.CreatedAt, &route.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.writeJSONError(w, "Route not found", http.StatusNotFound)
		} else {
			log.Printf("Error querying route: %v", err)
			s.writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Convert pq.StringArray to []string
	route.Methods = []string(methods)
	route.AuthKeys = []string(authKeys)

	writeJSON(w, route)
}

// handleAPIKeys handles API keys collection endpoints
func (s *Server) handleAPIKeys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getAPIKeys(w, r)
	case "POST":
		s.createAPIKey(w, r)
	}
}

// handleAPIKey handles individual API key endpoints
func (s *Server) handleAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeJSONError(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.getAPIKey(w, r, id)
	case "PUT":
		s.updateAPIKey(w, r, id)
	case "DELETE":
		s.deleteAPIKey(w, r, id)
	}
}

// getAPIKeys retrieves all API keys for the tenant
func (s *Server) getAPIKeys(w http.ResponseWriter, r *http.Request) {
	tenantID := s.getTenantID(r)

	keys, err := s.apiKeysService.GetByTenant(r.Context(), tenantID)
	if err != nil {
		log.Printf("Error getting API keys for tenant %d: %v", tenantID, err)
		s.writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format (don't include the hash)
	var response []map[string]interface{}
	for _, key := range keys {
		response = append(response, map[string]interface{}{
			"id":         key.ID,
			"name":       key.Name,
			"prefix":     key.Prefix, // Only show the prefix, not the full key
			"created_at": key.CreatedAt,
			"updated_at": key.UpdatedAt,
		})
	}

	writeJSON(w, response)
}

// createAPIKey creates a new API key
func (s *Server) createAPIKey(w http.ResponseWriter, r *http.Request) {
	tenantID := s.getTenantID(r)

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		s.writeJSONError(w, "Name is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement rate limiting by user/tenant here
	// For now, we'll skip rate limiting implementation

	// Create the API key
	apiKey := &models.APIKey{
		TenantID: tenantID,
		Name:     req.Name,
	}

	fullKey, err := s.apiKeysService.Create(r.Context(), apiKey)
	if err != nil {
		log.Printf("Error creating API key for tenant %d: %v", tenantID, err)
		s.writeJSONError(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	// Return the response with full key (only returned once)
	response := CreateAPIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       fullKey, // Full key returned only once
		Prefix:    apiKey.Prefix,
		CreatedAt: apiKey.CreatedAt,
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, response)
}

// updateAPIKey updates an existing API key
func (s *Server) updateAPIKey(w http.ResponseWriter, r *http.Request, id int) {
	var key APIKey
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		s.writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert to PostgreSQL array
	permissionsArray := pq.StringArray(key.Permissions)

	query := `
		UPDATE api_keys SET
			key_value = $1, name = $2, permissions = $3, rate_limit = $4, enabled = $5, expires_at = $6,
			updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`

	err := s.db.QueryRow(query,
		key.KeyValue, key.Name, permissionsArray, key.RateLimit, key.Enabled, key.ExpiresAt, id,
	).Scan(&key.UpdatedAt)

	if err != nil {
		log.Printf("Error updating API key: %v", err)
		s.writeJSONError(w, "Failed to update API key", http.StatusInternalServerError)
		return
	}

	key.ID = id
	writeJSON(w, key)
}

// deleteAPIKey deletes an API key
func (s *Server) deleteAPIKey(w http.ResponseWriter, r *http.Request, id int) {
	tenantID := s.getTenantID(r)

	err := s.apiKeysService.Delete(r.Context(), id, tenantID)
	if err != nil {
		log.Printf("Error deleting API key %d for tenant %d: %v", id, tenantID, err)
		s.writeJSONError(w, "Failed to delete API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getAPIKey retrieves a single API key
func (s *Server) getAPIKey(w http.ResponseWriter, r *http.Request, id int) {
	query := `
		SELECT id, key_value, name, permissions, rate_limit, enabled, expires_at, created_at, updated_at
		FROM api_keys
		WHERE id = $1
	`

	var key APIKey
	var permissions pq.StringArray

	err := s.db.QueryRow(query, id).Scan(
		&key.ID, &key.KeyValue, &key.Name, &permissions, &key.RateLimit,
		&key.Enabled, &key.ExpiresAt, &key.CreatedAt, &key.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.writeJSONError(w, "API key not found", http.StatusNotFound)
		} else {
			log.Printf("Error querying API key: %v", err)
			s.writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Convert pq.StringArray to []string
	key.Permissions = []string(permissions)

	writeJSON(w, key)
}

// handleDashboardStats handles dashboard statistics
func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	// Get active routes count
	var activeRoutes int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM proxy_routes WHERE enabled = true").Scan(&activeRoutes); err != nil {
		activeRoutes = 0 // Default to 0 if query fails
	}

	// Get active API keys count
	var activeAPIKeys int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE enabled = true").Scan(&activeAPIKeys); err != nil {
		activeAPIKeys = 0 // Default to 0 if query fails
	}

	// Mock other stats for now (in real implementation, these would come from metrics)
	stats := DashboardStats{
		TotalRequests:   2250,
		AvgResponseTime: 45.0,
		SuccessRate:     99.2,
		CacheHitRate:    68.5,
		ActiveRoutes:    activeRoutes,
		ActiveAPIKeys:   activeAPIKeys,
	}

	writeJSON(w, stats)
}

// handleDashboardActivity handles dashboard activity feed
func (s *Server) handleDashboardActivity(w http.ResponseWriter, r *http.Request) {
	// Mock activity data
	activity := []map[string]interface{}{
		{
			"id":        1,
			"message":   "New API key \"demo-client\" generated",
			"timestamp": "2 minutes ago",
			"type":      "success",
		},
		{
			"id":        2,
			"message":   "Route /api/v1/posts cache hit rate improved to 85%",
			"timestamp": "15 minutes ago",
			"type":      "success",
		},
		{
			"id":        3,
			"message":   "Rate limit exceeded for IP 192.168.1.100",
			"timestamp": "1 hour ago",
			"type":      "warning",
		},
	}

	writeJSON(w, activity)
}

// handleHealth handles health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Start starts the admin API server
func (s *Server) Start(addr string) error {
	log.Printf("Starting admin API server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// Close closes the database connection
func (s *Server) Close() error {
	return s.db.Close()
}

// validateURL validates if a string is a valid URL
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include a scheme (http:// or https://)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// validateAuthMode validates the auth_mode field
func validateAuthMode(mode string) error {
	validModes := []string{"none", "api_key", "basic"}
	for _, valid := range validModes {
		if mode == valid {
			return nil
		}
	}
	return fmt.Errorf("auth_mode must be one of: %s", strings.Join(validModes, ", "))
}

// validateJSON validates that a string contains valid JSON
func validateJSON(jsonStr string) error {
	if jsonStr == "" {
		return nil // Empty JSON is allowed
	}

	var temp interface{}
	if err := json.Unmarshal([]byte(jsonStr), &temp); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}
	return nil
}

// validateRouteInput validates the input for creating/updating routes
func (s *Server) validateRouteInput(route *models.Route) error {
	if route.Name == "" {
		return fmt.Errorf("name is required")
	}

	if route.MatchPath == "" {
		return fmt.Errorf("match_path is required")
	}

	if err := validateURL(route.UpstreamURL); err != nil {
		return fmt.Errorf("upstream_url validation failed: %w", err)
	}

	if err := validateAuthMode(route.AuthMode); err != nil {
		return err
	}

	// Validate JSON fields
	if len(route.HeadersJSON) > 0 {
		if err := validateJSON(string(route.HeadersJSON)); err != nil {
			return fmt.Errorf("headers_json validation failed: %w", err)
		}
	}

	if len(route.CachingPolicyJSON) > 0 {
		if err := validateJSON(string(route.CachingPolicyJSON)); err != nil {
			return fmt.Errorf("caching_policy_json validation failed: %w", err)
		}
	}

	if len(route.RateLimitPolicyJSON) > 0 {
		if err := validateJSON(string(route.RateLimitPolicyJSON)); err != nil {
			return fmt.Errorf("rate_limit_policy_json validation failed: %w", err)
		}
	}

	return nil
}

// handleV1Routes handles v1 routes collection endpoints with validation
func (s *Server) handleV1Routes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.getV1Routes(w, r)
	case "POST":
		s.createV1Route(w, r)
	}
}

// handleV1Route handles v1 individual route endpoints with validation
func (s *Server) handleV1Route(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.writeJSONError(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.getV1Route(w, r, id)
	case "PUT":
		s.updateV1Route(w, r, id)
	case "DELETE":
		s.deleteV1Route(w, r, id)
	}
}

// getV1Routes retrieves all routes using the new v1 format
func (s *Server) getV1Routes(w http.ResponseWriter, r *http.Request) {
	tenantID := s.getTenantID(r)

	routes, err := s.routesService.GetByTenant(r.Context(), tenantID)
	if err != nil {
		log.Printf("Error getting routes for tenant %d: %v", tenantID, err)
		s.writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, routes)
}

// createV1Route creates a new route with validation
func (s *Server) createV1Route(w http.ResponseWriter, r *http.Request) {
	tenantID := s.getTenantID(r)

	var route models.Route
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		s.writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set tenant ID
	route.TenantID = tenantID

	// Validate the route input
	if err := s.validateRouteInput(&route); err != nil {
		s.writeJSONError(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Create the route
	if err := s.routesService.Create(r.Context(), &route); err != nil {
		log.Printf("Error creating route for tenant %d: %v", tenantID, err)
		s.writeJSONError(w, "Failed to create route", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, route)
}

// updateV1Route updates an existing route with validation
func (s *Server) updateV1Route(w http.ResponseWriter, r *http.Request, id int) {
	tenantID := s.getTenantID(r)

	// First, get the existing route to ensure it belongs to the tenant
	existingRoute, err := s.routesService.GetByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.writeJSONError(w, "Route not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting route %d: %v", id, err)
			s.writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if existingRoute.TenantID != tenantID {
		s.writeJSONError(w, "Route not found", http.StatusNotFound)
		return
	}

	var route models.Route
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		s.writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set required fields
	route.ID = id
	route.TenantID = tenantID

	// Validate the route input
	if err := s.validateRouteInput(&route); err != nil {
		s.writeJSONError(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Update the route
	if err := s.routesService.Update(r.Context(), &route); err != nil {
		log.Printf("Error updating route %d for tenant %d: %v", id, tenantID, err)
		s.writeJSONError(w, "Failed to update route", http.StatusInternalServerError)
		return
	}

	writeJSON(w, route)
}

// deleteV1Route deletes a route
func (s *Server) deleteV1Route(w http.ResponseWriter, r *http.Request, id int) {
	tenantID := s.getTenantID(r)

	// Verify the route belongs to the tenant before deleting
	existingRoute, err := s.routesService.GetByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.writeJSONError(w, "Route not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting route %d: %v", id, err)
			s.writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if existingRoute.TenantID != tenantID {
		s.writeJSONError(w, "Route not found", http.StatusNotFound)
		return
	}

	if err := s.routesService.Delete(r.Context(), id, tenantID); err != nil {
		log.Printf("Error deleting route %d for tenant %d: %v", id, tenantID, err)
		s.writeJSONError(w, "Failed to delete route", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getV1Route retrieves a single route
func (s *Server) getV1Route(w http.ResponseWriter, r *http.Request, id int) {
	tenantID := s.getTenantID(r)

	route, err := s.routesService.GetByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.writeJSONError(w, "Route not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting route %d: %v", id, err)
			s.writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Verify the route belongs to the tenant
	if route.TenantID != tenantID {
		s.writeJSONError(w, "Route not found", http.StatusNotFound)
		return
	}

	writeJSON(w, route)
}
