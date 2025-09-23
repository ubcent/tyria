package apikeys

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/ubcent/edge.link/internal/models"
	_ "modernc.org/sqlite"
)

func TestAPIKeyGeneration(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create table
	_, err = db.Exec(`
		CREATE TABLE api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			name VARCHAR(255) NOT NULL,
			prefix VARCHAR(20) NOT NULL,
			hash VARCHAR(255) NOT NULL,
			last_used_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	service := NewService(db)

	t.Run("CreateAPIKey", func(t *testing.T) {
		apiKey := &models.APIKey{
			TenantID: 1,
			Name:     "test-key",
		}

		fullKey, err := service.Create(context.Background(), apiKey)
		if err != nil {
			t.Fatalf("Failed to create API key: %v", err)
		}

		// Verify key format (prefix.key)
		if !strings.Contains(fullKey, ".") {
			t.Errorf("Expected key format 'prefix.key', got: %s", fullKey)
		}

		// Verify prefix starts with "el_"
		if !strings.HasPrefix(fullKey, "el_") {
			t.Errorf("Expected key to start with 'el_', got: %s", fullKey)
		}

		// Verify prefix matches extracted prefix
		dotIndex := strings.Index(fullKey, ".")
		expectedPrefix := fullKey[:dotIndex]
		if apiKey.Prefix != expectedPrefix {
			t.Errorf("Expected prefix %s, got %s", expectedPrefix, apiKey.Prefix)
		}

		// Verify API key was stored with hashed key
		if apiKey.Hash == "" {
			t.Error("Expected hashed key to be stored")
		}

		// Verify full key is not stored in database
		if strings.Contains(apiKey.Hash, fullKey) {
			t.Error("Full key should not be stored in database")
		}
	})

	t.Run("ValidateAPIKey", func(t *testing.T) {
		// Create a key first
		apiKey := &models.APIKey{
			TenantID: 1,
			Name:     "validation-test-key",
		}

		fullKey, err := service.Create(context.Background(), apiKey)
		if err != nil {
			t.Fatalf("Failed to create API key: %v", err)
		}

		// Validate the key
		validatedKey, err := service.ValidateKey(context.Background(), fullKey)
		if err != nil {
			t.Fatalf("Failed to validate API key: %v", err)
		}

		if validatedKey.ID != apiKey.ID {
			t.Errorf("Expected key ID %d, got %d", apiKey.ID, validatedKey.ID)
		}

		if validatedKey.Name != apiKey.Name {
			t.Errorf("Expected key name %s, got %s", apiKey.Name, validatedKey.Name)
		}
	})

	t.Run("GetByTenant", func(t *testing.T) {
		// Create multiple keys for different tenants
		key1 := &models.APIKey{TenantID: 1, Name: "tenant1-key1"}
		key2 := &models.APIKey{TenantID: 1, Name: "tenant1-key2"}
		key3 := &models.APIKey{TenantID: 2, Name: "tenant2-key1"}

		_, err := service.Create(context.Background(), key1)
		if err != nil {
			t.Fatalf("Failed to create key1: %v", err)
		}
		_, err = service.Create(context.Background(), key2)
		if err != nil {
			t.Fatalf("Failed to create key2: %v", err)
		}
		_, err = service.Create(context.Background(), key3)
		if err != nil {
			t.Fatalf("Failed to create key3: %v", err)
		}

		// Get keys for tenant 1
		keys, err := service.GetByTenant(context.Background(), 1)
		if err != nil {
			t.Fatalf("Failed to get keys for tenant 1: %v", err)
		}

		if len(keys) != 4 { // Including the key from previous tests
			t.Errorf("Expected 4 keys for tenant 1, got %d", len(keys))
		}

		// Get keys for tenant 2
		keys, err = service.GetByTenant(context.Background(), 2)
		if err != nil {
			t.Fatalf("Failed to get keys for tenant 2: %v", err)
		}

		if len(keys) != 1 {
			t.Errorf("Expected 1 key for tenant 2, got %d", len(keys))
		}
	})

	t.Run("DeleteAPIKey", func(t *testing.T) {
		// Create a key to delete
		apiKey := &models.APIKey{
			TenantID: 1,
			Name:     "delete-test-key",
		}

		_, err := service.Create(context.Background(), apiKey)
		if err != nil {
			t.Fatalf("Failed to create API key: %v", err)
		}

		// Delete the key
		err = service.Delete(context.Background(), apiKey.ID, apiKey.TenantID)
		if err != nil {
			t.Fatalf("Failed to delete API key: %v", err)
		}

		// Verify key is deleted
		keys, err := service.GetByTenant(context.Background(), apiKey.TenantID)
		if err != nil {
			t.Fatalf("Failed to get keys: %v", err)
		}

		for _, key := range keys {
			if key.ID == apiKey.ID {
				t.Error("Key should have been deleted")
			}
		}
	})
}