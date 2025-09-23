package logging

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
)

func TestLoggingWithContext(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Override the output by creating logger manually for testing
	logger := Logger{
		Logger: zerolog.New(&buf).Level(zerolog.InfoLevel).With().Timestamp().Logger(),
	}

	// Test structured logging with all required fields
	contextLogger := logger.WithContext(123, 456, "req-123-abc", "sk_test_")

	contextLogger.Info().Msg("test message")

	// Parse the JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify all required fields are present
	expectedFields := map[string]interface{}{
		"tenant_id":       float64(123), // JSON unmarshals numbers as float64
		"route_id":        float64(456),
		"request_id":      "req-123-abc",
		"api_key_prefix":  "sk_test_",
		"message":         "test message",
	}

	for field, expectedValue := range expectedFields {
		if actualValue, exists := logEntry[field]; !exists {
			t.Errorf("Expected field %s not found in log entry", field)
		} else if actualValue != expectedValue {
			t.Errorf("Expected %s=%v, got %v", field, expectedValue, actualValue)
		}
	}

	// Verify timestamp exists
	if _, exists := logEntry["time"]; !exists {
		t.Error("Expected timestamp field 'time' not found in log entry")
	}
}

func TestRequestIDGeneration(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("Generated request ID should not be empty")
	}

	if id1 == id2 {
		t.Error("Generated request IDs should be unique")
	}

	// Basic UUID format validation (8-4-4-4-12 pattern)
	if len(id1) != 36 {
		t.Errorf("Expected UUID length 36, got %d", len(id1))
	}
}

func TestLoggingFormats(t *testing.T) {
	testCases := []struct {
		name   string
		format string
	}{
		{"JSON format", "json"},
		{"Console format", "console"},
		{"Text format", "text"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				Level:  "info",
				Format: tc.format,
				Output: "stdout",
			}

			logger, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			if logger == nil {
				t.Error("Logger should not be nil")
			}
		})
	}
}