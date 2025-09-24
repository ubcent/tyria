package customdomains

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/ubcent/edge.link/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create the custom_domains table
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

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestCustomDomainsService(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	service := NewService(db)
	ctx := context.Background()

	t.Run("Create custom domain", func(t *testing.T) {
		domain := &models.CustomDomain{
			TenantID: 1,
			Hostname: "api.example.com",
			Status:   "pending",
		}

		err := service.Create(ctx, domain)
		if err != nil {
			t.Errorf("Failed to create custom domain: %v", err)
		}

		if domain.ID == 0 {
			t.Errorf("Expected domain ID to be set")
		}

		if domain.VerificationToken == "" {
			t.Errorf("Expected verification token to be generated")
		}
	})

	t.Run("Get by tenant", func(t *testing.T) {
		domains, err := service.GetByTenant(ctx, 1)
		if err != nil {
			t.Errorf("Failed to get domains by tenant: %v", err)
		}

		if len(domains) != 1 {
			t.Errorf("Expected 1 domain, got %d", len(domains))
		}

		if domains[0].Hostname != "api.example.com" {
			t.Errorf("Expected hostname 'api.example.com', got %s", domains[0].Hostname)
		}
	})

	t.Run("Get by hostname", func(t *testing.T) {
		domain, err := service.GetByHostname(ctx, "api.example.com")
		if err != nil {
			t.Errorf("Failed to get domain by hostname: %v", err)
		}

		if domain.Hostname != "api.example.com" {
			t.Errorf("Expected hostname 'api.example.com', got %s", domain.Hostname)
		}

		if domain.TenantID != 1 {
			t.Errorf("Expected tenant ID 1, got %d", domain.TenantID)
		}
	})

	t.Run("Update status", func(t *testing.T) {
		err := service.UpdateStatus(ctx, 1, 1, "verified")
		if err != nil {
			t.Errorf("Failed to update domain status: %v", err)
		}

		// Verify status was updated
		domain, err := service.GetByHostname(ctx, "api.example.com")
		if err != nil {
			t.Errorf("Failed to get domain: %v", err)
		}

		if domain.Status != "verified" {
			t.Errorf("Expected status 'verified', got %s", domain.Status)
		}
	})

	t.Run("Delete custom domain", func(t *testing.T) {
		err := service.Delete(ctx, 1, 1)
		if err != nil {
			t.Errorf("Failed to delete custom domain: %v", err)
		}

		// Verify domain was deleted
		domains, err := service.GetByTenant(ctx, 1)
		if err != nil {
			t.Errorf("Failed to get domains by tenant: %v", err)
		}

		if len(domains) != 0 {
			t.Errorf("Expected 0 domains after deletion, got %d", len(domains))
		}
	})
}

func TestVerificationTokenGeneration(t *testing.T) {
	service := &Service{}

	token1, err := service.generateVerificationToken()
	if err != nil {
		t.Errorf("Failed to generate token: %v", err)
	}

	token2, err := service.generateVerificationToken()
	if err != nil {
		t.Errorf("Failed to generate token: %v", err)
	}

	if token1 == token2 {
		t.Errorf("Expected different tokens, got same: %s", token1)
	}

	if len(token1) == 0 {
		t.Errorf("Expected non-empty token")
	}
}
