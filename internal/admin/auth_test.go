package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthServerJSONErrors(t *testing.T) {
	// Create a minimal auth server for testing
	authServer := &AuthServer{}

	t.Run("Invalid signin request returns JSON", func(t *testing.T) {
		// Test with invalid JSON body
		reqBody := bytes.NewBufferString(`{"email":"test@example.com"}`) // missing password
		req := httptest.NewRequest("POST", "/api/auth/signin", reqBody)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		authServer.HandleSignin(w, req)

		// Check status code
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
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

		if errorResp.Error != "Bad Request" {
			t.Errorf("Expected error 'Bad Request', got '%s'", errorResp.Error)
		}

		if errorResp.Message != "Email and password are required" {
			t.Errorf("Expected message 'Email and password are required', got '%s'", errorResp.Message)
		}

		if errorResp.Code != 400 {
			t.Errorf("Expected code 400, got %d", errorResp.Code)
		}
	})

	t.Run("Invalid JSON body returns JSON error", func(t *testing.T) {
		// Test with malformed JSON
		reqBody := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest("POST", "/api/auth/signin", reqBody)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		authServer.HandleSignin(w, req)

		// Check status code
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
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

		if errorResp.Error != "Bad Request" {
			t.Errorf("Expected error 'Bad Request', got '%s'", errorResp.Error)
		}

		if errorResp.Message != "Invalid request body" {
			t.Errorf("Expected message 'Invalid request body', got '%s'", errorResp.Message)
		}

		if errorResp.Code != 400 {
			t.Errorf("Expected code 400, got %d", errorResp.Code)
		}
	})

	t.Run("Wrong HTTP method returns JSON error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/auth/signin", nil)
		w := httptest.NewRecorder()

		authServer.HandleSignin(w, req)

		// Check status code
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
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

		if errorResp.Error != "Method Not Allowed" {
			t.Errorf("Expected error 'Method Not Allowed', got '%s'", errorResp.Error)
		}

		if errorResp.Message != "Method not allowed" {
			t.Errorf("Expected message 'Method not allowed', got '%s'", errorResp.Message)
		}

		if errorResp.Code != 405 {
			t.Errorf("Expected code 405, got %d", errorResp.Code)
		}
	})
}