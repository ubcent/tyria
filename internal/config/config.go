package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration for the proxy service
type Config struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Cache    CacheConfig    `yaml:"cache" json:"cache"`
	Routes   []RouteConfig  `yaml:"routes" json:"routes"`
	APIKeys  []APIKeyConfig `yaml:"api_keys" json:"api_keys"`
	Logging  LoggingConfig  `yaml:"logging" json:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics" json:"metrics"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Host         string        `yaml:"host" json:"host"`
	Port         int           `yaml:"port" json:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
}

// CacheConfig holds cache-specific configuration
type CacheConfig struct {
	DefaultTTL    time.Duration `yaml:"default_ttl" json:"default_ttl"`
	MaxSize       int64         `yaml:"max_size" json:"max_size"`
	CleanupPeriod time.Duration `yaml:"cleanup_period" json:"cleanup_period"`
}

// RouteConfig defines a proxy route configuration
type RouteConfig struct {
	Path         string               `yaml:"path" json:"path"`
	Target       string               `yaml:"target" json:"target"`
	Methods      []string             `yaml:"methods" json:"methods"`
	Cache        RouteCacheConfig     `yaml:"cache" json:"cache"`
	RateLimit    RouteRateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
	Auth         RouteAuthConfig      `yaml:"auth" json:"auth"`
	Validation   RouteValidationConfig `yaml:"validation" json:"validation"`
}

// RouteCacheConfig holds route-specific cache configuration
type RouteCacheConfig struct {
	Enabled bool          `yaml:"enabled" json:"enabled"`
	TTL     time.Duration `yaml:"ttl" json:"ttl"`
}

// RouteRateLimitConfig holds route-specific rate limiting configuration
type RouteRateLimitConfig struct {
	Enabled   bool          `yaml:"enabled" json:"enabled"`
	Rate      int           `yaml:"rate" json:"rate"`
	Burst     int           `yaml:"burst" json:"burst"`
	Period    time.Duration `yaml:"period" json:"period"`
	PerClient bool          `yaml:"per_client" json:"per_client"`
}

// RouteAuthConfig holds route-specific authentication configuration
type RouteAuthConfig struct {
	Required bool     `yaml:"required" json:"required"`
	Keys     []string `yaml:"keys" json:"keys"`
}

// RouteValidationConfig holds route-specific validation configuration
type RouteValidationConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	RequestSchema string `yaml:"request_schema" json:"request_schema"`
	ResponseSchema string `yaml:"response_schema" json:"response_schema"`
}

// APIKeyConfig defines an API key configuration
type APIKeyConfig struct {
	Key         string   `yaml:"key" json:"key"`
	Name        string   `yaml:"name" json:"name"`
	Permissions []string `yaml:"permissions" json:"permissions"`
	RateLimit   int      `yaml:"rate_limit" json:"rate_limit"`
	Enabled     bool     `yaml:"enabled" json:"enabled"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	Output string `yaml:"output" json:"output"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path" json:"path"`
	Port    int    `yaml:"port" json:"port"`
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	
	// Try YAML first, then JSON
	if err := yaml.Unmarshal(data, &config); err != nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Set defaults
	config.setDefaults()
	
	return &config, nil
}

// setDefaults sets default values for the configuration
func (c *Config) setDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "localhost"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 120 * time.Second
	}

	if c.Cache.DefaultTTL == 0 {
		c.Cache.DefaultTTL = 5 * time.Minute
	}
	if c.Cache.MaxSize == 0 {
		c.Cache.MaxSize = 100 * 1024 * 1024 // 100MB
	}
	if c.Cache.CleanupPeriod == 0 {
		c.Cache.CleanupPeriod = 10 * time.Minute
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}

	if c.Metrics.Path == "" {
		c.Metrics.Path = "/metrics"
	}
	if c.Metrics.Port == 0 {
		c.Metrics.Port = 9090
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Metrics.Port < 1 || c.Metrics.Port > 65535 {
		return fmt.Errorf("invalid metrics port: %d", c.Metrics.Port)
	}

	for i, route := range c.Routes {
		if route.Path == "" {
			return fmt.Errorf("route %d: path cannot be empty", i)
		}
		if route.Target == "" {
			return fmt.Errorf("route %d: target cannot be empty", i)
		}
	}

	for i, key := range c.APIKeys {
		if key.Key == "" {
			return fmt.Errorf("api key %d: key cannot be empty", i)
		}
		if key.Name == "" {
			return fmt.Errorf("api key %d: name cannot be empty", i)
		}
	}

	return nil
}