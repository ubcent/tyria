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

	"github.com/ubcent/edge.link/internal/config"
	"github.com/ubcent/edge.link/internal/proxy"
)

const (
	defaultConfigPath = "config.yaml"
	version           = "1.0.0"
)

func main() {
	var (
		configPath = flag.String("config", defaultConfigPath, "Path to configuration file")
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

	// Print loaded configuration (excluding sensitive data)
	printConfig(cfg)

	// Create proxy service
	proxyService := proxy.New(cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      proxyService.Handler(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Create metrics server if enabled
	var metricsServer *http.Server
	if cfg.Metrics.Enabled {
		metricsMux := http.NewServeMux()
		metricsMux.HandleFunc(cfg.Metrics.Path, func(w http.ResponseWriter, r *http.Request) {
			stats := proxyService.GetMetrics().GetStats()
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(stats); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
		
		metricsServer = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Metrics.Port),
			Handler: metricsMux,
		}
	}

	// Start servers
	go func() {
		log.Printf("Starting proxy server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Proxy server failed: %v", err)
		}
	}()

	if metricsServer != nil {
		go func() {
			log.Printf("Starting metrics server on %s", metricsServer.Addr)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Metrics server failed: %v", err)
			}
		}()
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown proxy server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Proxy server shutdown error: %v", err)
	}

	// Shutdown metrics server
	if metricsServer != nil {
		if err := metricsServer.Shutdown(ctx); err != nil {
			log.Printf("Metrics server shutdown error: %v", err)
		}
	}

	log.Println("Servers stopped")
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

	return os.WriteFile(path, data, 0644)
}