// Package logging provides centralized logging functionality for the edge.link proxy service.
// It supports multiple output formats and destinations including file and console output.
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Config holds logging configuration
type Config struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	Output string `yaml:"output" json:"output"`
}

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	// Parse log level
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
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

	// Configure handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text", "":
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	logger := slog.New(handler)
	return &Logger{Logger: logger}, nil
}

// WithTenant adds tenant context to log entries
func (l *Logger) WithTenant(tenantID int) *Logger {
	return &Logger{
		Logger: l.With("tenant_id", tenantID),
	}
}

// WithRequest adds request context to log entries
func (l *Logger) WithRequest(method, path, clientIP string) *Logger {
	return &Logger{
		Logger: l.With(
			"method", method,
			"path", path,
			"client_ip", clientIP,
		),
	}
}

// WithRoute adds route context to log entries
func (l *Logger) WithRoute(routePath string) *Logger {
	return &Logger{
		Logger: l.With("route_path", routePath),
	}
}
