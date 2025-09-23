// Package metrics provides in-memory collection and exposition of proxy metrics.
package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Metrics collects and tracks various proxy metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests       int64 `json:"total_requests"`
	ProxiedRequests     int64 `json:"proxied_requests"`
	CachedRequests      int64 `json:"cached_requests"`
	FailedRequests      int64 `json:"failed_requests"`
	RateLimitedRequests int64 `json:"rate_limited_requests"`

	// Response metrics
	TotalResponseTime time.Duration `json:"total_response_time_ns"`

	// Route metrics
	RouteMetrics map[string]*RouteMetrics `json:"route_metrics"`

	// Status code metrics
	StatusCodes map[int]int64 `json:"status_codes"`

	// Start time
	StartTime time.Time `json:"start_time"`
}

// RouteMetrics tracks metrics for a specific route
type RouteMetrics struct {
	Requests     int64         `json:"requests"`
	CacheHits    int64         `json:"cache_hits"`
	CacheMisses  int64         `json:"cache_misses"`
	Errors       int64         `json:"errors"`
	ResponseTime time.Duration `json:"response_time_ns"`
	LastAccessed time.Time     `json:"last_accessed"`
}

// New creates a new metrics instance
func New() *Metrics {
	return &Metrics{
		RouteMetrics: make(map[string]*RouteMetrics),
		StatusCodes:  make(map[int]int64),
		StartTime:    time.Now(),
	}
}

// IncrementRequests increments the total request counter
func (m *Metrics) IncrementRequests() {
	m.mu.Lock()
	m.TotalRequests++
	m.mu.Unlock()
}

// IncrementProxiedRequests increments the proxied request counter
func (m *Metrics) IncrementProxiedRequests() {
	m.mu.Lock()
	m.ProxiedRequests++
	m.mu.Unlock()
}

// IncrementCachedRequests increments the cached request counter
func (m *Metrics) IncrementCachedRequests() {
	m.mu.Lock()
	m.CachedRequests++
	m.mu.Unlock()
}

// IncrementFailedRequests increments the failed request counter
func (m *Metrics) IncrementFailedRequests() {
	m.mu.Lock()
	m.FailedRequests++
	m.mu.Unlock()
}

// IncrementRateLimitedRequests increments the rate limited request counter
func (m *Metrics) IncrementRateLimitedRequests() {
	m.mu.Lock()
	m.RateLimitedRequests++
	m.mu.Unlock()
}

// RecordResponseTime records the response time for a request
func (m *Metrics) RecordResponseTime(duration time.Duration) {
	m.mu.Lock()
	m.TotalResponseTime += duration
	m.mu.Unlock()
}

// RecordStatusCode records a status code
func (m *Metrics) RecordStatusCode(code int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StatusCodes[code]++
}

// RecordRouteMetric records metrics for a specific route
func (m *Metrics) RecordRouteMetric(route string, hit bool, duration time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	routeMetric, exists := m.RouteMetrics[route]
	if !exists {
		routeMetric = &RouteMetrics{}
		m.RouteMetrics[route] = routeMetric
	}

	routeMetric.Requests++
	routeMetric.ResponseTime += duration
	routeMetric.LastAccessed = time.Now()

	if hit {
		routeMetric.CacheHits++
	} else {
		routeMetric.CacheMisses++
	}

	if isError {
		routeMetric.Errors++
	}
}

// GetStats returns a snapshot of current metrics
func (m *Metrics) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalRequests := m.TotalRequests
	proxiedRequests := m.ProxiedRequests
	cachedRequests := m.CachedRequests
	failedRequests := m.FailedRequests
	rateLimitedRequests := m.RateLimitedRequests
	totalResponseTime := m.TotalResponseTime

	var avgResponseTime time.Duration
	if totalRequests > 0 {
		avgResponseTime = totalResponseTime / time.Duration(totalRequests)
	}

	// Copy route metrics
	routeMetrics := make(map[string]RouteStats)
	for route, metrics := range m.RouteMetrics {
		var avgRouteResponseTime time.Duration
		if metrics.Requests > 0 {
			avgRouteResponseTime = metrics.ResponseTime / time.Duration(metrics.Requests)
		}

		routeMetrics[route] = RouteStats{
			Requests:        metrics.Requests,
			CacheHits:       metrics.CacheHits,
			CacheMisses:     metrics.CacheMisses,
			Errors:          metrics.Errors,
			AvgResponseTime: avgRouteResponseTime,
			LastAccessed:    metrics.LastAccessed,
		}
	}

	// Copy status codes
	statusCodes := make(map[int]int64)
	for code, count := range m.StatusCodes {
		statusCodes[code] = count
	}

	uptime := time.Since(m.StartTime)

	return Stats{
		TotalRequests:       totalRequests,
		ProxiedRequests:     proxiedRequests,
		CachedRequests:      cachedRequests,
		FailedRequests:      failedRequests,
		RateLimitedRequests: rateLimitedRequests,
		AvgResponseTime:     avgResponseTime,
		RouteMetrics:        routeMetrics,
		StatusCodes:         statusCodes,
		Uptime:              uptime,
		StartTime:           m.StartTime,
	}
}

// Stats represents a snapshot of metrics
type Stats struct {
	TotalRequests       int64                 `json:"total_requests"`
	ProxiedRequests     int64                 `json:"proxied_requests"`
	CachedRequests      int64                 `json:"cached_requests"`
	FailedRequests      int64                 `json:"failed_requests"`
	RateLimitedRequests int64                 `json:"rate_limited_requests"`
	AvgResponseTime     time.Duration         `json:"avg_response_time_ns"`
	RouteMetrics        map[string]RouteStats `json:"route_metrics"`
	StatusCodes         map[int]int64         `json:"status_codes"`
	Uptime              time.Duration         `json:"uptime_ns"`
	StartTime           time.Time             `json:"start_time"`
}

// RouteStats represents statistics for a specific route
type RouteStats struct {
	Requests        int64         `json:"requests"`
	CacheHits       int64         `json:"cache_hits"`
	CacheMisses     int64         `json:"cache_misses"`
	Errors          int64         `json:"errors"`
	AvgResponseTime time.Duration `json:"avg_response_time_ns"`
	LastAccessed    time.Time     `json:"last_accessed"`
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests = 0
	m.ProxiedRequests = 0
	m.CachedRequests = 0
	m.FailedRequests = 0
	m.RateLimitedRequests = 0
	m.TotalResponseTime = 0

	m.RouteMetrics = make(map[string]*RouteMetrics)
	m.StatusCodes = make(map[int]int64)
	m.StartTime = time.Now()
}

// Handler returns an HTTP handler that serves metrics as JSON
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		stats := m.GetStats()
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Middleware returns middleware that records request metrics
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment total requests
		m.IncrementRequests()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start)
		m.RecordResponseTime(duration)
		m.RecordStatusCode(wrapped.statusCode)

		// Record error if status code indicates failure
		if wrapped.statusCode >= 400 {
			m.IncrementFailedRequests()
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
