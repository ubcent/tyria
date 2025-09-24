// Package metrics provides Prometheus metrics collection and exposition for the edge.link proxy service.
package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics contains all Prometheus metrics for the proxy service
type PrometheusMetrics struct {
	// Counter: proxy_requests_total by tenant, route, status_code, cache_status
	requestsTotal *prometheus.CounterVec

	// Histogram: proxy_latency_ms
	latencyHistogram *prometheus.HistogramVec

	// Gauges for current tokens in rate limiter (optional)
	rateLimiterTokens *prometheus.GaugeVec

	// Registry for this instance
	registry *prometheus.Registry
}

// NewPrometheusMetrics creates a new instance of Prometheus metrics
func NewPrometheusMetrics() *PrometheusMetrics {
	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_requests_total",
			Help: "Total number of proxy requests by tenant, route, status code and cache status",
		},
		[]string{"tenant", "route", "status_code", "cache_status"},
	)

	latencyHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxy_latency_ms",
			Help:    "Proxy request latency in milliseconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tenant", "route"},
	)

	rateLimiterTokens := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "proxy_rate_limiter_tokens",
			Help: "Current number of available tokens in rate limiter",
		},
		[]string{"tenant", "route", "client_id"},
	)

	// Register all metrics with the registry
	registry.MustRegister(requestsTotal)
	registry.MustRegister(latencyHistogram)
	registry.MustRegister(rateLimiterTokens)

	return &PrometheusMetrics{
		requestsTotal:     requestsTotal,
		latencyHistogram:  latencyHistogram,
		rateLimiterTokens: rateLimiterTokens,
		registry:          registry,
	}
}

// GetRegistry returns the Prometheus registry for this metrics instance
func (pm *PrometheusMetrics) GetRegistry() *prometheus.Registry {
	return pm.registry
}

// RecordRequest records a request with all necessary labels
func (pm *PrometheusMetrics) RecordRequest(tenant, route string, statusCode int, cacheStatus string, latencyMs float64) {
	statusCodeStr := strconv.Itoa(statusCode)

	// Increment request counter
	pm.requestsTotal.WithLabelValues(tenant, route, statusCodeStr, cacheStatus).Inc()

	// Record latency
	pm.latencyHistogram.WithLabelValues(tenant, route).Observe(latencyMs)
}

// UpdateRateLimiterTokens updates the gauge for rate limiter tokens
func (pm *PrometheusMetrics) UpdateRateLimiterTokens(tenant, route, clientID string, tokens float64) {
	pm.rateLimiterTokens.WithLabelValues(tenant, route, clientID).Set(tokens)
}

// GetCacheStatus determines cache status from hit/miss boolean
func GetCacheStatus(hit bool) string {
	if hit {
		return "hit"
	}
	return "miss"
}

// DurationToMs converts time.Duration to milliseconds as float64
func DurationToMs(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}
