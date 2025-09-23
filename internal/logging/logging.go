// Package logging provides centralized logging functionality for the edge.link proxy service.
// It supports multiple output formats and destinations including file and console output.
package logging

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Config holds logging configuration
type Config struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	Output string `yaml:"output" json:"output"`
}

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	zerolog.Logger
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	// Parse log level
	var level zerolog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn", "warning":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		level = zerolog.InfoLevel
	}

	// Configure output
	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stdout", "":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, err
		}
		output = file
	}

	// Configure logger based on format
	var logger zerolog.Logger
	switch strings.ToLower(cfg.Format) {
	case "json":
		logger = zerolog.New(output).Level(level).With().Timestamp().Logger()
	case "text", "console", "":
		// Use console format for text output
		consoleOutput := zerolog.ConsoleWriter{Out: output}
		logger = zerolog.New(consoleOutput).Level(level).With().Timestamp().Logger()
	default:
		logger = zerolog.New(output).Level(level).With().Timestamp().Logger()
	}

	return &Logger{Logger: logger}, nil
}

// WithTenant adds tenant context to log entries
func (l *Logger) WithTenant(tenantID int) *Logger {
	return &Logger{
		Logger: l.With().Int("tenant_id", tenantID).Logger(),
	}
}

// WithRequest adds request context to log entries
func (l *Logger) WithRequest(method, path, clientIP string) *Logger {
	return &Logger{
		Logger: l.With().
			Str("method", method).
			Str("path", path).
			Str("client_ip", clientIP).
			Logger(),
	}
}

// WithRoute adds route context to log entries (updated to include route_id)
func (l *Logger) WithRoute(routeID int, routePath string) *Logger {
	return &Logger{
		Logger: l.With().
			Int("route_id", routeID).
			Str("route_path", routePath).
			Logger(),
	}
}

// WithAPIKey adds API key context to log entries
func (l *Logger) WithAPIKey(apiKeyPrefix string) *Logger {
	return &Logger{
		Logger: l.With().Str("api_key_prefix", apiKeyPrefix).Logger(),
	}
}

// WithRequestID adds request ID context to log entries
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.With().Str("request_id", requestID).Logger(),
	}
}

// WithContext creates a logger with multiple context fields at once
func (l *Logger) WithContext(tenantID int, routeID int, requestID, apiKeyPrefix string) *Logger {
	event := l.With().
		Int("tenant_id", tenantID).
		Str("request_id", requestID)

	if routeID > 0 {
		event = event.Int("route_id", routeID)
	}

	if apiKeyPrefix != "" {
		event = event.Str("api_key_prefix", apiKeyPrefix)
	}

	return &Logger{Logger: event.Logger()}
}
