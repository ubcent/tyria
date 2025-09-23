package proxy

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
	"github.com/ubcent/edge.link/internal/cache"
)

func TestSimpleDBProxy(t *testing.T) {
	// Create in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create minimal test cache
	testCache := cache.NewLRU(1024*1024, 5*60, 10*60)
	defer testCache.Stop()
	
	service := NewDBServiceWithCache(db, testCache)

	t.Run("health check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()

		service.healthHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		t.Logf("Health response: %s", w.Body.String())
	})
	
	t.Run("cache stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/cache/stats", nil)
		w := httptest.NewRecorder()

		service.cacheStatsHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		t.Logf("Cache stats response: %s", w.Body.String())
	})
}