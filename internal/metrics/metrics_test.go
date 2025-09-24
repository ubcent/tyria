package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusMetrics(t *testing.T) {
	// Create a new metrics instance
	metrics := New()

	// Record some test metrics
	metrics.RecordRouteMetricWithTenant("tenant1", "/api/v1/test", true, 50*time.Millisecond, false, 200)
	metrics.RecordRouteMetricWithTenant("tenant1", "/api/v1/test", false, 100*time.Millisecond, false, 200)
	metrics.RecordRouteMetricWithTenant("tenant2", "/api/v2/test", false, 200*time.Millisecond, true, 500)

	// Update rate limiter tokens
	metrics.UpdateRateLimiterTokens("tenant1", "/api/v1/test", "client1", 50)

	// Test that Prometheus metrics were recorded
	reg := metrics.prometheus.GetRegistry()

	// Check if we have the expected metrics families
	metricFamilies, err := reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	expectedMetrics := []string{
		"proxy_requests_total",
		"proxy_latency_ms",
		"proxy_rate_limiter_tokens",
	}

	for _, expectedMetric := range expectedMetrics {
		found := false
		for _, mf := range metricFamilies {
			if mf.GetName() == expectedMetric {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected metric %s not found", expectedMetric)
		}
	}

	// Test specific metric values - check individual label combinations
	tenant1Hit := testutil.ToFloat64(metrics.prometheus.requestsTotal.WithLabelValues("tenant1", "/api/v1/test", "200", "hit"))
	tenant1Miss := testutil.ToFloat64(metrics.prometheus.requestsTotal.WithLabelValues("tenant1", "/api/v1/test", "200", "miss"))
	tenant2Miss := testutil.ToFloat64(metrics.prometheus.requestsTotal.WithLabelValues("tenant2", "/api/v2/test", "500", "miss"))

	if tenant1Hit != 1.0 {
		t.Errorf("Expected tenant1 hit counter value 1.0, got %f", tenant1Hit)
	}
	if tenant1Miss != 1.0 {
		t.Errorf("Expected tenant1 miss counter value 1.0, got %f", tenant1Miss)
	}
	if tenant2Miss != 1.0 {
		t.Errorf("Expected tenant2 miss counter value 1.0, got %f", tenant2Miss)
	}

	// Test that labels are working
	expectedTenant1Counter := 2.0 // tenant1 had 2 requests
	actualTenant1Counter := tenant1Hit + tenant1Miss
	if actualTenant1Counter != expectedTenant1Counter {
		t.Errorf("Expected tenant1 counter value %f, got %f", expectedTenant1Counter, actualTenant1Counter)
	}

	// Test rate limiter gauge
	expectedGaugeValue := 50.0
	actualGaugeValue := testutil.ToFloat64(metrics.prometheus.rateLimiterTokens.WithLabelValues("tenant1", "/api/v1/test", "client1"))
	if actualGaugeValue != expectedGaugeValue {
		t.Errorf("Expected gauge value %f, got %f", expectedGaugeValue, actualGaugeValue)
	}
}

func TestPrometheusHandler(t *testing.T) {
	metrics := New()

	// Record a test metric
	metrics.RecordRouteMetricWithTenant("test_tenant", "/test", false, 100*time.Millisecond, false, 200)

	// Get the Prometheus handler
	handler := metrics.PrometheusHandler()
	if handler == nil {
		t.Fatal("PrometheusHandler returned nil")
	}

	// The handler should be ready to serve Prometheus format metrics
	// We can't easily test the HTTP output here, but we can verify it doesn't panic
}

func TestCacheStatus(t *testing.T) {
	tests := []struct {
		hit      bool
		expected string
	}{
		{true, "hit"},
		{false, "miss"},
	}

	for _, test := range tests {
		result := GetCacheStatus(test.hit)
		if result != test.expected {
			t.Errorf("GetCacheStatus(%t) = %s, expected %s", test.hit, result, test.expected)
		}
	}
}

func TestDurationToMs(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected float64
	}{
		{time.Millisecond, 1.0},
		{100 * time.Millisecond, 100.0},
		{time.Second, 1000.0},
		{500 * time.Microsecond, 0.5},
	}

	for _, test := range tests {
		result := DurationToMs(test.duration)
		if result != test.expected {
			t.Errorf("DurationToMs(%v) = %f, expected %f", test.duration, result, test.expected)
		}
	}
}
