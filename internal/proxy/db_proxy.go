// Package proxy provides core proxy functionality for the edge.link service.
// It handles request routing, caching, rate limiting, and response processing.
package proxy

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ubcent/edge.link/internal/apikeys"
	"github.com/ubcent/edge.link/internal/cache"
	"github.com/ubcent/edge.link/internal/customdomains"
	"github.com/ubcent/edge.link/internal/metrics"
	"github.com/ubcent/edge.link/internal/models"
	"github.com/ubcent/edge.link/internal/ratelimit"
	"github.com/ubcent/edge.link/internal/requestlogs"
	"github.com/ubcent/edge.link/internal/routes"
	"github.com/ubcent/edge.link/internal/tenant"
)

// Cache configuration constants
const (
	defaultCacheSizeMB  = 100
	defaultCacheTTL     = 5 * time.Minute
	defaultCacheCleanup = 10 * time.Minute
	httpErrorStatusCode = 400
	bytesToMB           = 1024 * 1024
)

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
	Message    string
	RetryAfter int
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	return e.Message
}

// DBService represents a database-driven proxy service
type DBService struct {
	db             *sql.DB
	routesService  *routes.Service
	tenantService  *tenant.Service
	apiKeysService *apikeys.Service
	domainsService *customdomains.Service
	logsService    *requestlogs.Service
	cache          cache.Interface
	keyBuilder     *cache.KeyBuilder
	limiter        *ratelimit.Service
	metrics        *metrics.Metrics
}

// NewDBService creates a new database-driven proxy service
func NewDBService(db *sql.DB) *DBService {
	// Initialize LRU cache as default
	cacheInstance := cache.NewLRU(defaultCacheSizeMB*bytesToMB, defaultCacheTTL, defaultCacheCleanup) // 100MB cache, 5min TTL, 10min cleanup

	// Initialize in-memory rate limiter as default
	rateLimitServiceConfig := ratelimit.ServiceConfig{
		UseRedis: false,
		InMemoryConfig: ratelimit.Config{
			MaxTokens:     100,
			RefillRate:    10,
			RefillPeriod:  time.Second,
			CleanupPeriod: 10 * time.Minute,
		},
	}
	limiterInstance := ratelimit.NewService(rateLimitServiceConfig)

	return &DBService{
		db:             db,
		routesService:  routes.NewService(db),
		tenantService:  tenant.NewService(db),
		apiKeysService: apikeys.NewService(db),
		domainsService: customdomains.NewService(db),
		logsService:    requestlogs.NewService(db),
		cache:          cacheInstance,
		keyBuilder:     cache.NewKeyBuilder(),
		limiter:        limiterInstance,
		metrics:        metrics.New(),
	}
}

// NewDBServiceWithCache creates a new database-driven proxy service with custom cache
func NewDBServiceWithCache(db *sql.DB, cacheImpl cache.Interface) *DBService {
	// Initialize in-memory rate limiter as default
	rateLimitServiceConfig := ratelimit.ServiceConfig{
		UseRedis: false,
		InMemoryConfig: ratelimit.Config{
			MaxTokens:     100,
			RefillRate:    10,
			RefillPeriod:  time.Second,
			CleanupPeriod: 10 * time.Minute,
		},
	}
	limiterInstance := ratelimit.NewService(rateLimitServiceConfig)

	return &DBService{
		db:             db,
		routesService:  routes.NewService(db),
		tenantService:  tenant.NewService(db),
		apiKeysService: apikeys.NewService(db),
		domainsService: customdomains.NewService(db),
		logsService:    requestlogs.NewService(db),
		cache:          cacheImpl,
		keyBuilder:     cache.NewKeyBuilder(),
		limiter:        limiterInstance,
		metrics:        metrics.New(),
	}
}

// NewDBServiceWithRateLimit creates a new database-driven proxy service with custom cache and rate limiting
func NewDBServiceWithRateLimit(db *sql.DB, cacheImpl cache.Interface, rateLimitConfig ratelimit.ServiceConfig) *DBService {
	limiterInstance := ratelimit.NewService(rateLimitConfig)

	return &DBService{
		db:             db,
		routesService:  routes.NewService(db),
		tenantService:  tenant.NewService(db),
		apiKeysService: apikeys.NewService(db),
		domainsService: customdomains.NewService(db),
		logsService:    requestlogs.NewService(db),
		cache:          cacheImpl,
		keyBuilder:     cache.NewKeyBuilder(),
		limiter:        limiterInstance,
		metrics:        metrics.New(),
	}
}

// Handler returns the main HTTP handler for the database-driven proxy
func (s *DBService) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/api/health", s.healthHandler)

	// Cache management endpoints
	mux.HandleFunc("/api/cache/stats", s.cacheStatsHandler)
	mux.HandleFunc("/api/cache/clear", s.cacheClearHandler)

	// Proxy handler with tenant extraction
	mux.HandleFunc("/", s.dbProxyHandler)

	return mux
}

// dbProxyHandler handles proxy requests with database-driven routing
func (s *DBService) dbProxyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	cacheStatus := cache.CacheStatusMiss

	// Resolve tenant from hostname or X-Tenant header
	tenantID, err := s.resolveTenant(r)
	if err != nil {
		http.Error(w, "Unable to resolve tenant", http.StatusBadRequest)
		return
	}

	// For dev mode, handle /:tenant/* pattern
	if tenantID == 0 {
		tenantID, err = s.extractTenantFromPath(r)
		if err != nil {
			http.Error(w, "Invalid tenant in path", http.StatusBadRequest)
			return
		}
	}

	if tenantID == 0 {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	// Find matching route
	route, pathParams, err := s.findMatchingRoute(r, tenantID)
	if err != nil {
		http.Error(w, "Route resolution error", http.StatusInternalServerError)
		return
	}

	if route == nil {
		http.Error(w, "No route found", http.StatusNotFound)
		s.logRequest(tenantID, nil, r, start, http.StatusNotFound, 0, 0, string(cache.CacheStatusMiss))
		return
	}

	// Enforce authentication
	if err := s.enforceAuth(route, r, tenantID); err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "forbidden" {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		s.logRequest(tenantID, &route.ID, r, start, status, 0, 0, string(cache.CacheStatusMiss))
		return
	}

	// Enforce rate limiting
	if err := s.enforceRateLimit(route, r, tenantID); err != nil {
		status := http.StatusTooManyRequests
		retryAfter := ""

		// Extract retry-after from error if available
		if rateLimitErr, ok := err.(*RateLimitError); ok {
			retryAfter = strconv.Itoa(rateLimitErr.RetryAfter)
		}

		w.Header().Set("Retry-After", retryAfter)
		http.Error(w, err.Error(), status)
		s.logRequest(tenantID, &route.ID, r, start, status, 0, 0, string(cache.CacheStatusMiss))
		return
	}

	// Check cache if enabled and request is cacheable
	var cacheKey string
	var cachingPolicy *models.CachingPolicy

	if len(route.CachingPolicyJSON) > 0 {
		cachingPolicy, err = route.GetCachingPolicy()
		if err == nil && cachingPolicy.Enabled && cache.IsCacheable(r.Method) {
			// Generate cache key
			varyHeaders := s.extractVaryHeaders(r, cachingPolicy.VaryHeaders)

			if r.Method == "GET" || r.Method == "HEAD" {
				cacheKey = s.keyBuilder.GenerateKey(tenantID, route.Name, r.Method, r.URL.Path, r.URL.RawQuery, varyHeaders)
			} else {
				// For non-GET/HEAD, read body for cache key (though they aren't cached, this prepares for future)
				requestBody, err := io.ReadAll(r.Body)
				if err == nil {
					r.Body = io.NopCloser(bytes.NewReader(requestBody))
					cacheKey = s.keyBuilder.GenerateKeyWithBody(tenantID, route.Name, r.Method, r.URL.Path, r.URL.RawQuery, requestBody, varyHeaders)
				}
			}

			// Try to get from cache
			if cacheKey != "" {
				if cachedData, found := s.cache.Get(cacheKey); found {
					w.Header().Set("X-Cache-Status", string(cache.CacheStatusHit))
					w.Header().Set("Content-Type", "application/json")
					if _, err := w.Write(cachedData); err != nil {
						http.Error(w, "Failed to write cached response", http.StatusInternalServerError)
						return
					}
					cacheStatus = cache.CacheStatusHit

					// Log cache hit
					s.logRequest(tenantID, &route.ID, r, start, http.StatusOK, s.calculateRequestSize(r), int64(len(cachedData)), string(cacheStatus))
					return
				}
			}
		} else {
			cacheStatus = cache.CacheStatusBypass
		}
	} else {
		cacheStatus = cache.CacheStatusBypass
	}

	// Forward request
	bytesIn := s.calculateRequestSize(r)
	responseStatus, bytesOut, responseData := s.forwardRequestWithCaching(w, r, route, pathParams, cachingPolicy, cacheKey)

	// Cache successful responses if caching is enabled
	if cachingPolicy != nil && cachingPolicy.Enabled && cacheKey != "" &&
		cache.IsCacheable(r.Method) && responseStatus >= 200 && responseStatus < 300 &&
		len(responseData) > 0 {
		ttl := time.Duration(cachingPolicy.TTLSeconds) * time.Second
		if ttl == 0 {
			ttl = defaultCacheTTL // Default TTL
		}
		s.cache.SetWithTTL(cacheKey, responseData, ttl)
	}

	// Set cache status header
	w.Header().Set("X-Cache-Status", string(cacheStatus))

	// Log request
	s.logRequest(tenantID, &route.ID, r, start, responseStatus, bytesIn, bytesOut, string(cacheStatus))
}

// resolveTenant resolves tenant ID from hostname (custom domain) or X-Tenant header
func (s *DBService) resolveTenant(r *http.Request) (int, error) {
	// First check X-Tenant header for dev mode
	if tenantHeader := r.Header.Get("X-Tenant"); tenantHeader != "" {
		tenantID, err := strconv.Atoi(tenantHeader)
		if err != nil {
			return 0, fmt.Errorf("invalid X-Tenant header: %v", err)
		}
		return tenantID, nil
	}

	// Check for custom domain
	hostname := r.Host
	if idx := strings.Index(hostname, ":"); idx != -1 {
		hostname = hostname[:idx] // Remove port
	}

	domain, err := s.domainsService.GetByHostname(context.Background(), hostname)
	if err != nil {
		// Not found is not an error, just means no custom domain
		if strings.Contains(err.Error(), "not found") {
			return 0, nil
		}
		return 0, fmt.Errorf("error checking custom domain: %v", err)
	}

	return domain.TenantID, nil
}

// extractTenantFromPath extracts tenant ID from /:tenant/* path pattern
func (s *DBService) extractTenantFromPath(r *http.Request) (int, error) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 1 || pathParts[0] == "" {
		return 0, fmt.Errorf("no tenant in path")
	}

	// Try to parse as tenant ID
	tenantID, err := strconv.Atoi(pathParts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid tenant ID in path: %v", err)
	}

	// Remove tenant from path for further processing
	if len(pathParts) > 1 {
		r.URL.Path = "/" + strings.Join(pathParts[1:], "/")
	} else {
		r.URL.Path = "/"
	}

	return tenantID, nil
}

// findMatchingRoute finds a matching route for the request
func (s *DBService) findMatchingRoute(r *http.Request, tenantID int) (*models.Route, map[string]string, error) {
	routes, err := s.routesService.GetByTenant(context.Background(), tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get routes: %v", err)
	}

	requestPath := r.URL.Path

	for _, route := range routes {
		if !route.Enabled {
			continue
		}

		// Try to match the path
		if params, matches := s.matchPath(route.MatchPath, requestPath); matches {
			return route, params, nil
		}
	}

	return nil, nil, nil
}

// matchPath matches a route pattern against a request path, supporting path parameters and wildcards
func (s *DBService) matchPath(pattern, requestPath string) (map[string]string, bool) {
	params := make(map[string]string)

	// Handle exact match first
	if pattern == requestPath {
		return params, true
	}

	// Handle wildcard at the end
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		if strings.HasPrefix(requestPath, prefix) {
			return params, true
		}
	}

	// Handle path parameters like /users/{id} or /users/{id}/posts/{postId}
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")

	// Different number of parts means no match (unless wildcard)
	if len(patternParts) != len(requestParts) {
		return params, false
	}

	for i, patternPart := range patternParts {
		if i >= len(requestParts) {
			return params, false
		}

		requestPart := requestParts[i]

		// Check if this is a parameter
		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			paramName := patternPart[1 : len(patternPart)-1]
			params[paramName] = requestPart
			continue
		}

		// Exact match required for non-parameter parts
		if patternPart != requestPart {
			return params, false
		}
	}

	return params, true
}

// enforceAuth enforces authentication based on route auth_mode
func (s *DBService) enforceAuth(route *models.Route, r *http.Request, tenantID int) error {
	switch route.AuthMode {
	case "none":
		return nil
	case "api_key":
		return s.enforceAPIKeyAuth(route, r, tenantID)
	case "basic":
		return s.enforceBasicAuth(route, r, tenantID)
	default:
		return fmt.Errorf("unsupported auth mode: %s", route.AuthMode)
	}
}

// enforceAPIKeyAuth enforces API key authentication
func (s *DBService) enforceAPIKeyAuth(route *models.Route, r *http.Request, tenantID int) error {
	// Extract API key from various sources
	var apiKey string

	// Check Authorization header (Bearer token)
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			apiKey = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Check X-API-Key header
	if apiKey == "" {
		apiKey = r.Header.Get("X-API-Key")
	}

	// Check query parameter
	if apiKey == "" {
		apiKey = r.URL.Query().Get("api_key")
	}

	if apiKey == "" {
		return fmt.Errorf("API key required")
	}

	// Validate API key
	validatedKey, err := s.apiKeysService.ValidateKey(context.Background(), apiKey)
	if err != nil {
		return fmt.Errorf("invalid API key")
	}

	// Check if key belongs to the correct tenant
	if validatedKey.TenantID != tenantID {
		return fmt.Errorf("forbidden")
	}

	return nil
}

// enforceBasicAuth enforces basic authentication (placeholder - would need user service)
func (s *DBService) enforceBasicAuth(route *models.Route, r *http.Request, tenantID int) error {
	// TODO: Implement basic auth when user service is available
	return fmt.Errorf("basic authentication not implemented")
}

// forwardRequestWithCaching forwards the request to the upstream URL and captures response for caching
func (s *DBService) forwardRequestWithCaching(w http.ResponseWriter, r *http.Request, route *models.Route, pathParams map[string]string, cachingPolicy *models.CachingPolicy, cacheKey string) (int, int64, []byte) {
	// Substitute path parameters in upstream URL
	upstreamURLStr := route.UpstreamURL
	for paramName, paramValue := range pathParams {
		upstreamURLStr = strings.ReplaceAll(upstreamURLStr, "{"+paramName+"}", paramValue)
	}

	// Parse upstream URL
	upstreamURL, err := url.Parse(upstreamURLStr)
	if err != nil {
		http.Error(w, "Invalid upstream URL", http.StatusInternalServerError)
		return http.StatusInternalServerError, 0, nil
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)

	// Track response metrics and capture response data for caching
	responseWriter := &cachingResponseTracker{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		captureBody:    cachingPolicy != nil && cachingPolicy.Enabled && cache.IsCacheable(r.Method),
		buffer:         &bytes.Buffer{},
	}

	// Customize director to modify request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Apply header overrides
		headers, err := route.GetHeaders()
		if err == nil {
			for key, value := range headers {
				// Replace path parameters in header values
				for paramName, paramValue := range pathParams {
					value = strings.ReplaceAll(value, "{"+paramName+"}", paramValue)
				}
				req.Header.Set(key, value)
			}
		}

		// Set forwarded headers
		req.Header.Set("X-Forwarded-Host", r.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
		if r.Header.Get("X-Forwarded-For") == "" {
			req.Header.Set("X-Forwarded-For", s.getClientIP(r))
		}
	}

	// Execute proxy request
	proxy.ServeHTTP(responseWriter, r)

	responseData := responseWriter.buffer.Bytes()
	return responseWriter.statusCode, responseWriter.bytesWritten, responseData
}

// cachingResponseTracker tracks response metrics and captures response data for caching
type cachingResponseTracker struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	captureBody  bool
	buffer       *bytes.Buffer
}

func (crt *cachingResponseTracker) WriteHeader(statusCode int) {
	crt.statusCode = statusCode
	crt.ResponseWriter.WriteHeader(statusCode)
}

func (crt *cachingResponseTracker) Write(data []byte) (int, error) {
	n, err := crt.ResponseWriter.Write(data)
	crt.bytesWritten += int64(n)

	// Capture response data for caching if enabled and status is successful
	if crt.captureBody && crt.statusCode >= 200 && crt.statusCode < 300 {
		crt.buffer.Write(data)
	}

	return n, err
}

// responseTracker tracks response metrics
//
//nolint:unused // currently not used; kept for potential extension of metrics tracking
type responseTracker struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

//nolint:unused // currently not used
func (rt *responseTracker) WriteHeader(statusCode int) {
	rt.statusCode = statusCode
	rt.ResponseWriter.WriteHeader(statusCode)
}

//nolint:unused // currently not used
func (rt *responseTracker) Write(data []byte) (int, error) {
	n, err := rt.ResponseWriter.Write(data)
	rt.bytesWritten += int64(n)
	return n, err
}

// extractVaryHeaders extracts the specified vary headers from the request
func (s *DBService) extractVaryHeaders(r *http.Request, varyHeaders []string) map[string]string {
	headers := make(map[string]string)
	for _, headerName := range varyHeaders {
		if value := r.Header.Get(headerName); value != "" {
			headers[headerName] = value
		}
	}
	return headers
}

// calculateRequestSize calculates the size of the incoming request
func (s *DBService) calculateRequestSize(r *http.Request) int64 {
	size := int64(len(r.URL.Path) + len(r.URL.RawQuery))

	// Add headers size
	for name, values := range r.Header {
		size += int64(len(name))
		for _, value := range values {
			size += int64(len(value))
		}
	}

	// Add content length if available
	if r.ContentLength > 0 {
		size += r.ContentLength
	}

	return size
}

// logRequest logs the request to the database
func (s *DBService) logRequest(tenantID int, routeID *int, r *http.Request, start time.Time, statusCode int, bytesIn, bytesOut int64, cacheStatus string) {
	latencyMs := int(time.Since(start).Milliseconds())
	duration := time.Since(start)

	log := &models.RequestLog{
		TenantID:    tenantID,
		RouteID:     routeID,
		StatusCode:  statusCode,
		LatencyMs:   latencyMs,
		CacheStatus: cacheStatus,
		BytesIn:     int(bytesIn),
		BytesOut:    int(bytesOut),
		CreatedAt:   time.Now(),
	}

	// Record metrics for Prometheus
	tenantIDStr := fmt.Sprintf("%d", tenantID)
	route := "unknown"
	if routeID != nil {
		route = fmt.Sprintf("route_%d", *routeID)
	}

	// Determine if it's a cache hit
	cacheHit := cacheStatus == string(cache.CacheStatusHit)
	isError := statusCode >= 400

	// Record metrics (both JSON and Prometheus)
	s.metrics.RecordRouteMetricWithTenant(tenantIDStr, route, cacheHit, duration, isError, statusCode)

	// Log asynchronously to avoid blocking the request
	go func() {
		if err := s.logsService.Log(context.Background(), log); err != nil {
			// In a real implementation, you'd want proper error logging here
			fmt.Printf("Failed to log request: %v\n", err)
		}
	}()
}

// GetMetrics returns the metrics instance for the DB service
func (s *DBService) GetMetrics() *metrics.Metrics {
	return s.metrics
}

// getClientIP extracts the client IP address from the request
func (s *DBService) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// enforceRateLimit enforces rate limiting for a route and tenant
func (s *DBService) enforceRateLimit(route *models.Route, r *http.Request, tenantID int) error {
	// Parse rate limit policy from route
	if len(route.RateLimitPolicyJSON) == 0 {
		return nil // No rate limiting configured
	}

	rateLimitPolicy, err := route.GetRateLimitPolicy()
	if err != nil {
		// Log error but don't fail the request due to policy parsing issues
		return nil
	}

	if !rateLimitPolicy.Enabled {
		return nil // Rate limiting disabled
	}

	// Generate rate limit key for tenant+route
	tenantRouteKey := ratelimit.GenerateTenantRouteKey(tenantID, route.MatchPath)

	// Check tenant+route rate limit
	allowed, retryAfter := s.limiter.AllowWithPolicy(
		tenantRouteKey,
		rateLimitPolicy.RequestsPerMinute,
		rateLimitPolicy.Burst,
	)

	if !allowed {
		return &RateLimitError{
			Message:    fmt.Sprintf("Rate limit exceeded for tenant %d on route %s", tenantID, route.MatchPath),
			RetryAfter: retryAfter,
		}
	}

	// Check API key specific rate limiting if an API key is present
	apiKey := s.extractAPIKey(r)
	if apiKey != "" {
		// Validate the API key to check if it exists and get its info
		keyInfo, err := s.apiKeysService.ValidateKey(context.Background(), apiKey)
		if err == nil && keyInfo != nil && keyInfo.TenantID == tenantID {
			// Use API key prefix for rate limiting
			keyPrefix := keyInfo.Prefix
			if keyPrefix == "" && len(apiKey) > 8 {
				keyPrefix = apiKey[:8] // Use first 8 chars as fallback prefix
			}

			apiKeyRateLimitKey := ratelimit.GenerateAPIKeyKey(keyPrefix)

			// Use the same policy as the route for API key rate limiting
			// In a more advanced implementation, API keys could have their own policies
			allowed, retryAfter := s.limiter.AllowWithPolicy(
				apiKeyRateLimitKey,
				rateLimitPolicy.RequestsPerMinute,
				rateLimitPolicy.Burst,
			)

			if !allowed {
				return &RateLimitError{
					Message:    fmt.Sprintf("Rate limit exceeded for API key %s", keyPrefix),
					RetryAfter: retryAfter,
				}
			}
		}
	}

	return nil
}

// extractAPIKey extracts API key from request headers or query parameters
func (s *DBService) extractAPIKey(r *http.Request) string {
	// Check Authorization header for Bearer token
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// Check X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Check api_key query parameter
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// healthHandler returns health status
func (s *DBService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s", "service": "db-proxy"}`, time.Now().Format(time.RFC3339))
}

// cacheStatsHandler returns cache statistics
func (s *DBService) cacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.cache.Stats()
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Failed to encode cache stats", http.StatusInternalServerError)
	}
}

// cacheClearHandler clears the cache
func (s *DBService) cacheClearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.cache.Clear()
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"message": "Cache cleared successfully", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
}
