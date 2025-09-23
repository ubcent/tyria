// Package main provides the entry point for the edge.link proxy service.
// It initializes the server, loads configuration, and handles graceful shutdown.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ubcent/edge.link/internal/config"
	"github.com/ubcent/edge.link/internal/db"
	"github.com/ubcent/edge.link/internal/logging"
	"github.com/ubcent/edge.link/internal/proxy"
)

const (
	defaultConfigPath = "config.yaml"
	version           = "1.0.0"
)

func main() {
	var (
		configPath  = flag.String("config", defaultConfigPath, "Path to configuration file")
		showVersion = flag.Bool("version", false, "Show version and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("edge.link proxy v%s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize logging
	logger, err := logging.New(logging.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize database connection (optional for now)
	var database *db.DB
	if cfg.Database.Host != "" {
		database, err = db.New(db.Config{
			Host:            cfg.Database.Host,
			Port:            cfg.Database.Port,
			User:            cfg.Database.User,
			Password:        cfg.Database.Password,
			Database:        cfg.Database.Database,
			SSLMode:         cfg.Database.SSLMode,
			MaxOpenConns:    cfg.Database.MaxOpenConns,
			MaxIdleConns:    cfg.Database.MaxIdleConns,
			ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		})
		if err != nil {
			logger.Warn("Failed to connect to database, running without DB", "error", err)
		} else {
			defer func() { _ = database.Close() }()
		}
	}

	// Print loaded configuration (excluding sensitive data)
	printConfig(cfg)

	// Create chi router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Add health and version endpoints (database passed safely)
	r.Get("/healthz", healthHandler(database))
	r.Get("/version", versionHandler)

	// Create proxy service and mount its handler
	var proxyService *proxy.Service
	var dbProxyService *proxy.DBService

	if database != nil {
		// Use database-driven proxy when database is available
		dbProxyService = proxy.NewDBService(database.DB)
		r.Mount("/", dbProxyService.Handler())
	} else {
		// Fall back to config-driven proxy
		proxyService = proxy.New(cfg)
		r.Mount("/", proxyService.Handler())
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Create metrics server if enabled
	var metricsServer *http.Server
	if cfg.Metrics.Enabled {
		metricsRouter := chi.NewRouter()
		metricsRouter.Get(cfg.Metrics.Path, func(w http.ResponseWriter, _ *http.Request) {
			var stats interface{}
			if proxyService != nil {
				stats = proxyService.GetMetrics().GetStats()
			} else {
				// For DB proxy, return basic health info for now
				stats = map[string]interface{}{
					"service": "db-proxy",
					"status":  "healthy",
				}
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(stats); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})

		metricsServer = &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Metrics.Port),
			Handler:           metricsRouter,
			ReadHeaderTimeout: 10 * time.Second,
		}
	}

	// Start servers
	go func() {
		logger.Info("Starting proxy server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Proxy server failed", "error", err)
			os.Exit(1)
		}
	}()

	if metricsServer != nil {
		go func() {
			logger.Info("Starting metrics server", "addr", metricsServer.Addr)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("Metrics server failed", "error", err)
				os.Exit(1)
			}
		}()
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown proxy server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Proxy server shutdown error", "error", err)
	}

	// Shutdown metrics server
	if metricsServer != nil {
		if err := metricsServer.Shutdown(ctx); err != nil {
			logger.Error("Metrics server shutdown error", "error", err)
		}
	}

	logger.Info("Servers stopped")
}

// healthHandler returns health status including database connectivity
func healthHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   version,
		}

		// Check database connectivity if database is available
		if database != nil {
			if err := database.Health(); err != nil {
				health["status"] = "unhealthy"
				health["database"] = "unhealthy"
				health["error"] = err.Error()
				w.WriteHeader(http.StatusServiceUnavailable)
			} else {
				health["database"] = "healthy"
			}
		} else {
			health["database"] = "not_configured"
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(health); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// versionHandler returns version information
func versionHandler(w http.ResponseWriter, r *http.Request) {
	versionInfo := map[string]interface{}{
		"version":   version,
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "edge.link-proxy",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(versionInfo); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// printConfig prints the loaded configuration (excluding sensitive data)
func printConfig(cfg *config.Config) {
	log.Printf("Loaded configuration:")
	log.Printf("  Server: %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("  Cache: TTL=%v, MaxSize=%d bytes", cfg.Cache.DefaultTTL, cfg.Cache.MaxSize)
	log.Printf("  Routes: %d configured", len(cfg.Routes))
	log.Printf("  API Keys: %d configured", len(cfg.APIKeys))
	log.Printf("  Logging: Level=%s, Format=%s", cfg.Logging.Level, cfg.Logging.Format)
	log.Printf("  Metrics: Enabled=%t", cfg.Metrics.Enabled)

	// Print routes summary
	for i, route := range cfg.Routes {
		log.Printf("  Route %d: %s -> %s (methods: %v, cache: %t, auth: %t)",
			i+1, route.Path, route.Target, route.Methods,
			route.Cache.Enabled, route.Auth.Required)
	}
}

// createDefaultConfig creates a default configuration file if none exists
//
//nolint:unused // currently unused helper retained for potential CLI feature
func createDefaultConfig(path string) error {
	defaultCfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Cache: config.CacheConfig{
			DefaultTTL:    5 * time.Minute,
			MaxSize:       100 * 1024 * 1024, // 100MB
			CleanupPeriod: 10 * time.Minute,
		},
		Routes: []config.RouteConfig{
			{
				Path:    "/api/v1/",
				Target:  "https://jsonplaceholder.typicode.com",
				Methods: []string{"GET", "POST"},
				Cache: config.RouteCacheConfig{
					Enabled: true,
					TTL:     5 * time.Minute,
				},
				RateLimit: config.RouteRateLimitConfig{
					Enabled:   true,
					Rate:      100,
					Burst:     10,
					Period:    time.Minute,
					PerClient: true,
				},
				Auth: config.RouteAuthConfig{
					Required: false,
				},
			},
		},
		APIKeys: []config.APIKeyConfig{
			{
				Key:         "demo-key-12345",
				Name:        "demo",
				Permissions: []string{"proxy.*"},
				RateLimit:   1000,
				Enabled:     true,
			},
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Metrics: config.MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
			Port:    9090,
		},
	}

	data, err := json.MarshalIndent(defaultCfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
