// Package main provides the entry point for the edge.link admin API service.
// It initializes the admin server, database connections, and handles graceful shutdown.
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ubcent/edge.link/internal/admin"
	"github.com/ubcent/edge.link/internal/logging"
	"github.com/ubcent/edge.link/internal/tracing"
	_ "modernc.org/sqlite"
)

// Database constants
const (
	memoryDBURL = "file::memory:?cache=shared"
)

func main() {
	var (
		addr        = flag.String("addr", ":3001", "Admin API server address")
		databaseURL = flag.String("db", os.Getenv("POSTGRES_URL"), "PostgreSQL database URL")
		useSQLite   = flag.Bool("sqlite", false, "Use SQLite for testing")
	)
	flag.Parse()

	if *useSQLite || *databaseURL == "" {
		*databaseURL = memoryDBURL
		log.Println("Using in-memory SQLite database for testing")
	}

	// Initialize logging
	logger, err := logging.New(logging.Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize tracing
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "http://localhost:4318"
	}

	tracingProvider, err := tracing.NewProvider(tracing.Config{
		ServiceName:    "edgelink-admin-api",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		OTLPEndpoint:   otlpEndpoint,
		Enabled:        true,
	})
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to initialize tracing, continuing without tracing")
		tracingProvider = &tracing.Provider{} // Use no-op provider
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracingProvider.Shutdown(ctx); err != nil {
				logger.Error().Err(err).Msg("Failed to shutdown tracing provider")
			}
		}()
		logger.Info().Str("endpoint", otlpEndpoint).Msg("Tracing initialized")
	}

	// Create admin server
	server, err := admin.NewServer(*databaseURL)
	if err != nil {
		log.Fatalf("Failed to create admin server: %v", err)
	}
	defer func() { _ = server.Close() }()

	// Setup schema for SQLite if needed
	if *useSQLite || *databaseURL == "file::memory:?cache=shared" {
		if err := setupSQLiteSchema(server); err != nil {
			logger.Error().Err(err).Msg("Failed to setup SQLite schema")
			// Close server explicitly since we are about to exit and deferred calls won't run after os.Exit
			_ = server.Close()
			os.Exit(1)
		}
	}

	// Start server in goroutine
	go func() {
		if err := server.Start(*addr); err != nil {
			logger.Fatal().Err(err).Msg("Admin server failed")
		}
	}()

	logger.Info().Str("addr", *addr).Msg("Admin API server started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Admin API server stopped")
}

func setupSQLiteSchema(_ *admin.Server) error {
	// We need to access the database connection from the server
	// For now, let's create a simple schema using SQL statements
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	schema := `
	-- Tenants table for multi-tenant support
	CREATE TABLE IF NOT EXISTS tenants (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  name VARCHAR(255) NOT NULL,
	  plan VARCHAR(50) DEFAULT 'free' CHECK (plan IN ('free', 'pro', 'enterprise')),
	  status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'canceled')),
	  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Users table for authentication
	CREATE TABLE IF NOT EXISTS users (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
	  email VARCHAR(255) UNIQUE NOT NULL,
	  hashed_password VARCHAR(255) NOT NULL,
	  role VARCHAR(50) DEFAULT 'viewer' CHECK (role IN ('owner', 'admin', 'viewer')),
	  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- API keys management
	CREATE TABLE IF NOT EXISTS api_keys (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
	  name VARCHAR(255) NOT NULL,
	  prefix VARCHAR(20) NOT NULL,
	  hash VARCHAR(255) NOT NULL,
	  last_used_at DATETIME NULL,
	  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	  UNIQUE(tenant_id, name)
	);

	-- Proxy routes configuration  
	CREATE TABLE IF NOT EXISTS routes (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
	  name VARCHAR(255) NOT NULL,
	  match_path VARCHAR(255) NOT NULL,
	  upstream_url VARCHAR(500) NOT NULL,
	  headers_json TEXT DEFAULT '{}',
	  auth_mode VARCHAR(50) DEFAULT 'none' CHECK (auth_mode IN ('none', 'api_key', 'bearer')),
	  caching_policy_json TEXT DEFAULT '{"enabled": false, "ttl_seconds": 300}',
	  rate_limit_policy_json TEXT DEFAULT '{"enabled": false, "requests_per_minute": 100}',
	  enabled BOOLEAN DEFAULT 1,
	  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	  UNIQUE(tenant_id, match_path)
	);

	-- Custom domains for tenant branding
	CREATE TABLE IF NOT EXISTS custom_domains (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
	  hostname VARCHAR(255) NOT NULL UNIQUE,
	  verification_token VARCHAR(255) NOT NULL,
	  status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'failed')),
	  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Request logs for analytics and monitoring
	CREATE TABLE IF NOT EXISTS requests_log (
	  id INTEGER PRIMARY KEY AUTOINCREMENT,
	  tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
	  route_id INTEGER REFERENCES routes(id) ON DELETE SET NULL,
	  status_code INTEGER NOT NULL,
	  latency_ms INTEGER NOT NULL,
	  cache_status VARCHAR(20) DEFAULT 'miss' CHECK (cache_status IN ('hit', 'miss', 'bypass')),
	  bytes_in INTEGER DEFAULT 0,
	  bytes_out INTEGER DEFAULT 0,
	  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	return err
}
