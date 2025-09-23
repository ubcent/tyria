package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Config represents the main configuration for the proxy service
type Config struct {
	Server       ServerConfig       `yaml:"server" json:"server"`
	Database     DatabaseConfig     `yaml:"database" json:"database"`
	Cache        CacheConfig        `yaml:"cache" json:"cache"`
	Routes       []RouteConfig      `yaml:"routes" json:"routes"`
	APIKeys      []APIKeyConfig     `yaml:"api_keys" json:"api_keys"`
	Logging      LoggingConfig      `yaml:"logging" json:"logging"`
	Metrics      MetricsConfig      `yaml:"metrics" json:"metrics"`
	FeatureFlags FeatureFlagsConfig `yaml:"feature_flags" json:"feature_flags"`
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
	Path       string                `yaml:"path" json:"path"`
	Target     string                `yaml:"target" json:"target"`
	Methods    []string              `yaml:"methods" json:"methods"`
	Cache      RouteCacheConfig      `yaml:"cache" json:"cache"`
	RateLimit  RouteRateLimitConfig  `yaml:"rate_limit" json:"rate_limit"`
	Auth       RouteAuthConfig       `yaml:"auth" json:"auth"`
	Validation RouteValidationConfig `yaml:"validation" json:"validation"`
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
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	RequestSchema  string `yaml:"request_schema" json:"request_schema"`
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

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	User            string        `yaml:"user" json:"user"`
	Password        string        `yaml:"password" json:"password"`
	Database        string        `yaml:"database" json:"database"`
	SSLMode         string        `yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// FeatureFlagsConfig holds feature flag configuration
type FeatureFlagsConfig struct {
	MultiTenant       bool `yaml:"multi_tenant" json:"multi_tenant"`
	DomainLinking     bool `yaml:"domain_linking" json:"domain_linking"`
	UserAuth          bool `yaml:"user_auth" json:"user_auth"`
	BillingEnabled    bool `yaml:"billing_enabled" json:"billing_enabled"`
	EmailVerification bool `yaml:"email_verification" json:"email_verification"`
	SSLProvisioning   bool `yaml:"ssl_provisioning" json:"ssl_provisioning"`
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

	// Database defaults
	if c.Database.Host == "" {
		c.Database.Host = getEnvWithDefault("DB_HOST", "localhost")
	}
	if c.Database.Port == 0 {
		c.Database.Port = getEnvIntWithDefault("DB_PORT", 5432)
	}
	if c.Database.User == "" {
		c.Database.User = getEnvWithDefault("DB_USER", "edgelink")
	}
	if c.Database.Password == "" {
		c.Database.Password = getEnvWithDefault("DB_PASSWORD", "edgelink")
	}
	if c.Database.Database == "" {
		c.Database.Database = getEnvWithDefault("DB_NAME", "edgelink")
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = getEnvWithDefault("DB_SSL_MODE", "disable")
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = getEnvIntWithDefault("DB_MAX_OPEN_CONNS", 25)
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = getEnvIntWithDefault("DB_MAX_IDLE_CONNS", 5)
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 5 * time.Minute
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

	// Feature flags from environment variables
	c.FeatureFlags.MultiTenant = getEnvBoolWithDefault("FEATURE_MULTI_TENANT", false)
	c.FeatureFlags.DomainLinking = getEnvBoolWithDefault("FEATURE_DOMAIN_LINKING", false)
	c.FeatureFlags.UserAuth = getEnvBoolWithDefault("FEATURE_USER_AUTH", false)
	c.FeatureFlags.BillingEnabled = getEnvBoolWithDefault("FEATURE_BILLING", false)
	c.FeatureFlags.EmailVerification = getEnvBoolWithDefault("FEATURE_EMAIL_VERIFICATION", false)
	c.FeatureFlags.SSLProvisioning = getEnvBoolWithDefault("FEATURE_SSL_PROVISIONING", false)
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

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntWithDefault gets an environment variable as int with a default value
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolWithDefault gets an environment variable as bool with a default value
func getEnvBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
