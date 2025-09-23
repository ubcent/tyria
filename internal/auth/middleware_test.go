package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareJSONErrors(t *testing.T) {
	// Create JWT manager and middleware
	jwtManager := NewJWTManager("test-secret", "test-issuer")
	middleware := NewAuthMiddleware(jwtManager)

	t.Run("Missing authentication token returns JSON", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		protectedHandler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		protectedHandler.ServeHTTP(w, req)

		// Check status code
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		// Check content type
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}

		// Check JSON structure
		var errorResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
			Code    int    `json:"code"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if errorResp.Error != "Unauthorized" {
			t.Errorf("Expected error 'Unauthorized', got '%s'", errorResp.Error)
		}

		if errorResp.Message != "Missing authentication token" {
			t.Errorf("Expected message 'Missing authentication token', got '%s'", errorResp.Message)
		}

		if errorResp.Code != 401 {
			t.Errorf("Expected code 401, got %d", errorResp.Code)
		}
	})

	t.Run("Invalid authentication token returns JSON", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		protectedHandler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}))

		protectedHandler.ServeHTTP(w, req)

		// Check status code
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		// Check content type
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}

		// Check JSON structure
		var errorResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
			Code    int    `json:"code"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if errorResp.Error != "Unauthorized" {
			t.Errorf("Expected error 'Unauthorized', got '%s'", errorResp.Error)
		}

		if errorResp.Message != "Invalid authentication token" {
			t.Errorf("Expected message 'Invalid authentication token', got '%s'", errorResp.Message)
		}

		if errorResp.Code != 401 {
			t.Errorf("Expected code 401, got %d", errorResp.Code)
		}
	})

	t.Run("Insufficient permissions returns JSON", func(t *testing.T) {
		// Create a valid token with viewer role
		token, err := jwtManager.GenerateToken(1, 1, "test@example.com", "viewer")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Require admin role but user has viewer role
		adminHandler := middleware.RequireRole(RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("admin success"))
		}))

		// First authenticate, then check role
		authHandler := middleware.RequireAuth(adminHandler)
		authHandler.ServeHTTP(w, req)

		// Check status code
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}

		// Check content type
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}

		// Check JSON structure
		var errorResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
			Code    int    `json:"code"`
		}

		if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if errorResp.Error != "Forbidden" {
			t.Errorf("Expected error 'Forbidden', got '%s'", errorResp.Error)
		}

		if errorResp.Message != "Insufficient permissions" {
			t.Errorf("Expected message 'Insufficient permissions', got '%s'", errorResp.Message)
		}

		if errorResp.Code != 403 {
			t.Errorf("Expected code 403, got %d", errorResp.Code)
		}
	})
}
