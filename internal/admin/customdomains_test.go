package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ubcent/edge.link/internal/models"
)

func TestCustomDomainsEndpoints(t *testing.T) {
	// Create a test server with in-memory database
	server, err := NewServer("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer func() { _ = server.db.Close() }()

	// Setup schema
	schema := `
		CREATE TABLE custom_domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			hostname VARCHAR(255) NOT NULL UNIQUE,
			verification_token VARCHAR(255) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := server.db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	t.Run("Create custom domain", func(t *testing.T) {
		reqBody := map[string]string{
			"hostname": "api.example.com",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.handleCustomDomains(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var domain models.CustomDomain
		if err := json.NewDecoder(w.Body).Decode(&domain); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if domain.Hostname != "api.example.com" {
			t.Errorf("Expected hostname 'api.example.com', got %s", domain.Hostname)
		}

		if domain.VerificationToken == "" {
			t.Errorf("Expected verification token to be set")
		}

		if domain.Status != "pending" {
			t.Errorf("Expected status 'pending', got %s", domain.Status)
		}
	})

	t.Run("Get custom domains", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/domains", nil)
		w := httptest.NewRecorder()

		server.handleCustomDomains(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var domains []*models.CustomDomain
		if err := json.NewDecoder(w.Body).Decode(&domains); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if len(domains) != 1 {
			t.Errorf("Expected 1 domain, got %d", len(domains))
		}

		if domains[0].Hostname != "api.example.com" {
			t.Errorf("Expected hostname 'api.example.com', got %s", domains[0].Hostname)
		}
	})

	t.Run("Get specific custom domain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/domains/1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		w := httptest.NewRecorder()

		server.handleCustomDomain(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var domain models.CustomDomain
		if err := json.NewDecoder(w.Body).Decode(&domain); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if domain.Hostname != "api.example.com" {
			t.Errorf("Expected hostname 'api.example.com', got %s", domain.Hostname)
		}
	})

	t.Run("Verify custom domain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/domains/1/verify", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		w := httptest.NewRecorder()

		server.handleVerifyCustomDomain(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if _, ok := response["verified"]; !ok {
			t.Errorf("Expected 'verified' field in response")
		}

		if _, ok := response["status"]; !ok {
			t.Errorf("Expected 'status' field in response")
		}

		if _, ok := response["message"]; !ok {
			t.Errorf("Expected 'message' field in response")
		}
	})

	t.Run("Delete custom domain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/domains/1", nil)
		req = mux.SetURLVars(req, map[string]string{"id": "1"})
		w := httptest.NewRecorder()

		server.handleCustomDomain(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
		}

		// Verify domain was deleted
		req = httptest.NewRequest(http.MethodGet, "/api/v1/domains", nil)
		w = httptest.NewRecorder()
		server.handleCustomDomains(w, req)

		var domains []*models.CustomDomain
		if err := json.NewDecoder(w.Body).Decode(&domains); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if len(domains) != 0 {
			t.Errorf("Expected 0 domains after deletion, got %d", len(domains))
		}
	})
}

func TestWellKnownEndpoint(t *testing.T) {
	// Create a test server with in-memory database
	server, err := NewServer("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer func() { _ = server.db.Close() }()

	// Setup schema and test data
	schema := `
		CREATE TABLE custom_domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			hostname VARCHAR(255) NOT NULL UNIQUE,
			verification_token VARCHAR(255) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO custom_domains (tenant_id, hostname, verification_token, status)
		VALUES (1, 'api.example.com', 'test-token-123', 'pending');
	`
	if _, err := server.db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	t.Run("Well-known endpoint returns verification info", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/edge-link.txt", nil)
		req.Host = "api.example.com"
		w := httptest.NewRecorder()

		server.handleWellKnownEdgeLink(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		response := w.Body.String()
		if !contains(response, "edge-link verification") {
			t.Errorf("Expected response to contain 'edge-link verification'")
		}

		if !contains(response, "tenant_id: 1") {
			t.Errorf("Expected response to contain 'tenant_id: 1'")
		}

		if !contains(response, "verification_token: test-token-123") {
			t.Errorf("Expected response to contain verification token")
		}

		if !contains(response, "domain: api.example.com") {
			t.Errorf("Expected response to contain domain name")
		}
	})

	t.Run("Well-known endpoint returns 404 for unknown domain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/edge-link.txt", nil)
		req.Host = "unknown.example.com"
		w := httptest.NewRecorder()

		server.handleWellKnownEdgeLink(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d: %s", w.Code, w.Body.String())
		}

		response := w.Body.String()
		if !contains(response, "domain not configured") {
			t.Errorf("Expected response to contain 'domain not configured'")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
