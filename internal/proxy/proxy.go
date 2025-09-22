package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ubcent/edge.link/internal/auth"
	"github.com/ubcent/edge.link/internal/cache"
	"github.com/ubcent/edge.link/internal/config"
	"github.com/ubcent/edge.link/internal/metrics"
	"github.com/ubcent/edge.link/internal/ratelimit"
	"github.com/ubcent/edge.link/internal/validation"
)

// Service represents the proxy service
type Service struct {
	config    *config.Config
	cache     *cache.Cache
	auth      *auth.Manager
	limiter   *ratelimit.Limiter
	metrics   *metrics.Metrics
	validator *validation.Validator
}

// New creates a new proxy service
func New(cfg *config.Config) *Service {
	// Initialize cache
	cacheInstance := cache.New(cfg.Cache.MaxSize, cfg.Cache.DefaultTTL, cfg.Cache.CleanupPeriod)

	// Initialize rate limiter
	limiterConfig := ratelimit.Config{
		MaxTokens:     100,
		RefillRate:    10,
		RefillPeriod:  time.Second,
		CleanupPeriod: 10 * time.Minute,
	}
	limiterInstance := ratelimit.NewLimiter(limiterConfig)

	// Initialize auth manager
	authManager := auth.NewManager()
	for _, keyConfig := range cfg.APIKeys {
		authManager.AddKey(&auth.APIKey{
			Key:         keyConfig.Key,
			Name:        keyConfig.Name,
			Permissions: keyConfig.Permissions,
			RateLimit:   keyConfig.RateLimit,
			Enabled:     keyConfig.Enabled,
		})
	}

	return &Service{
		config:    cfg,
		cache:     cacheInstance,
		auth:      authManager,
		limiter:   limiterInstance,
		metrics:   metrics.New(),
		validator: validation.New(),
	}
}

// Handler returns the main HTTP handler for the proxy
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()

	// Management endpoints
	mux.HandleFunc("/api/health", s.healthHandler)
	mux.HandleFunc("/api/stats", s.statsHandler)
	mux.HandleFunc("/api/cache/stats", s.cacheStatsHandler)
	mux.HandleFunc("/api/cache/clear", s.cacheClearHandler)
	mux.HandleFunc("/api/auth/keys", s.authKeysHandler)
	mux.HandleFunc("/api/ratelimit/stats", s.rateLimitStatsHandler)

	// Proxy handler (catch-all)
	mux.HandleFunc("/", s.proxyHandler)

	// Wrap with metrics middleware
	return s.metrics.Middleware(mux)
}

// proxyHandler handles proxy requests
func (s *Service) proxyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Find matching route
	route := s.findRoute(r)
	if route == nil {
		http.Error(w, "No route found", http.StatusNotFound)
		s.metrics.RecordRouteMetric("unknown", false, time.Since(start), true)
		return
	}

	// Check authentication if required
	if route.Auth.Required {
		apiKey := auth.ExtractAPIKey(
			r.Header.Get("Authorization"),
			r.URL.Query().Get("api_key"),
			r.Header.Get("X-API-Key"),
		)

		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			s.metrics.RecordRouteMetric(route.Path, false, time.Since(start), true)
			return
		}

		// Validate API key and permissions
		validatedKey, err := s.auth.CheckPermission(apiKey, "proxy."+route.Path)
		if err != nil {
			status := http.StatusUnauthorized
			if err == auth.ErrPermissionDenied {
				status = http.StatusForbidden
			}
			http.Error(w, err.Error(), status)
			s.metrics.RecordRouteMetric(route.Path, false, time.Since(start), true)
			return
		}

		// Check per-key rate limiting
		if validatedKey.RateLimit > 0 {
			keyRateLimitKey := ratelimit.GenerateClientKey(validatedKey.Name)
			if !s.limiter.Allow(keyRateLimitKey) {
				http.Error(w, "Rate limit exceeded for API key", http.StatusTooManyRequests)
				s.metrics.IncrementRateLimitedRequests()
				s.metrics.RecordRouteMetric(route.Path, false, time.Since(start), true)
				return
			}
		}
	}

	// Check route-specific rate limiting
	if route.RateLimit.Enabled {
		var rateLimitKey string
		if route.RateLimit.PerClient {
			clientIP := getClientIP(r)
			rateLimitKey = ratelimit.GenerateClientEndpointKey(clientIP, route.Path)
		} else {
			rateLimitKey = ratelimit.GenerateEndpointKey(route.Path)
		}

		if !s.limiter.Allow(rateLimitKey) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			s.metrics.IncrementRateLimitedRequests()
			s.metrics.RecordRouteMetric(route.Path, false, time.Since(start), true)
			return
		}
	}

	// Check cache for GET requests
	cacheHit := false
	if route.Cache.Enabled && r.Method == "GET" {
		cacheKey := cache.GenerateKey(r.Method, r.URL.Path, r.URL.RawQuery)
		if cachedData, found := s.cache.Get(cacheKey); found {
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("Content-Type", "application/json")
			w.Write(cachedData)
			cacheHit = true
			s.metrics.IncrementCachedRequests()
			s.metrics.RecordRouteMetric(route.Path, true, time.Since(start), false)
			return
		}
	}

	// Validate request if enabled
	if route.Validation.Enabled && route.Validation.RequestSchema != "" {
		if !s.validator.HasSchema(route.Validation.RequestSchema) {
			http.Error(w, "Request validation schema not found", http.StatusInternalServerError)
			s.metrics.RecordRouteMetric(route.Path, false, time.Since(start), true)
			return
		}

		result, body, err := s.validator.ValidateRequest(route.Validation.RequestSchema, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Request validation error: %v", err), http.StatusInternalServerError)
			s.metrics.RecordRouteMetric(route.Path, false, time.Since(start), true)
			return
		}

		if !result.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			// Return validation errors
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "Request validation failed",
				"validation_errors": result.Errors,
			})
			s.metrics.RecordRouteMetric(route.Path, false, time.Since(start), true)
			return
		}

		// Restore body
		r.Body = io.NopCloser(bytes.NewReader(body))
	}

	// Proxy the request
	s.proxyRequest(w, r, route, start, cacheHit)
}

// proxyRequest performs the actual proxy operation
func (s *Service) proxyRequest(w http.ResponseWriter, r *http.Request, route *config.RouteConfig, start time.Time, cacheHit bool) {
	// Parse target URL
	target, err := url.Parse(route.Target)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		s.metrics.RecordRouteMetric(route.Path, false, time.Since(start), true)
		return
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Customize the director to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// Remove the route path prefix from the request URL
		req.URL.Path = strings.TrimPrefix(req.URL.Path, route.Path)
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		
		// Set headers
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
		if req.Header.Get("X-Forwarded-For") == "" {
			req.Header.Set("X-Forwarded-For", getClientIP(r))
		}
	}

	// Wrap response writer to capture response for caching
	wrapped := newCachingResponseWriter(w, route, s.cache, cache.GenerateKey(r.Method, r.URL.Path, r.URL.RawQuery))

	// Execute the proxy request
	proxy.ServeHTTP(wrapped, r)

	// Record metrics
	s.metrics.IncrementProxiedRequests()
	s.metrics.RecordRouteMetric(route.Path, cacheHit, time.Since(start), wrapped.statusCode >= 400)
}

// findRoute finds the matching route for a request
func (s *Service) findRoute(r *http.Request) *config.RouteConfig {
	for _, route := range s.config.Routes {
		// Check path match (simple prefix matching for MVP)
		if strings.HasPrefix(r.URL.Path, route.Path) {
			// Check method match
			if len(route.Methods) == 0 || contains(route.Methods, r.Method) {
				return &route
			}
		}
	}
	return nil
}

// cachingResponseWriter wraps http.ResponseWriter to capture responses for caching
type cachingResponseWriter struct {
	http.ResponseWriter
	buffer     *bytes.Buffer
	route      *config.RouteConfig
	cache      *cache.Cache
	cacheKey   string
	statusCode int
}

func newCachingResponseWriter(w http.ResponseWriter, route *config.RouteConfig, cache *cache.Cache, cacheKey string) *cachingResponseWriter {
	return &cachingResponseWriter{
		ResponseWriter: w,
		buffer:         &bytes.Buffer{},
		route:          route,
		cache:          cache,
		cacheKey:       cacheKey,
		statusCode:     200, // Default status code
	}
}

func (crw *cachingResponseWriter) Write(data []byte) (int, error) {
	// Write to buffer for caching if it's a cacheable response
	if crw.route.Cache.Enabled && crw.statusCode >= 200 && crw.statusCode < 300 {
		crw.buffer.Write(data)
	}
	n, err := crw.ResponseWriter.Write(data)
	
	// Cache the response when done writing
	if crw.route.Cache.Enabled && crw.statusCode >= 200 && crw.statusCode < 300 {
		go crw.writeToCache()
	}
	
	return n, err
}

func (crw *cachingResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
	crw.ResponseWriter.WriteHeader(statusCode)
}

func (crw *cachingResponseWriter) writeToCache() {
	if crw.route.Cache.Enabled && crw.statusCode >= 200 && crw.statusCode < 300 && crw.buffer.Len() > 0 {
		ttl := crw.route.Cache.TTL
		if ttl == 0 {
			ttl = 5 * time.Minute // Default TTL
		}
		crw.cache.SetWithTTL(crw.cacheKey, crw.buffer.Bytes(), ttl)
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetMetrics returns the metrics instance
func (s *Service) GetMetrics() *metrics.Metrics {
	return s.metrics
}

// Management handlers

func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
}

func (s *Service) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.metrics.GetStats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Service) cacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.cache.Stats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Service) cacheClearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.cache.Clear()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Cache cleared successfully"}`)
}

func (s *Service) authKeysHandler(w http.ResponseWriter, r *http.Request) {
	keys := s.auth.ListKeys()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(keys); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Service) rateLimitStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := s.limiter.Stats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}