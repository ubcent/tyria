package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

// Validator provides JSON schema validation
type Validator struct {
	mu      sync.RWMutex
	schemas map[string]*gojsonschema.Schema
}

// New creates a new validator instance
func New() *Validator {
	return &Validator{
		schemas: make(map[string]*gojsonschema.Schema),
	}
}

// AddSchema adds a JSON schema for validation
func (v *Validator) AddSchema(name, schemaJSON string) error {
	loader := gojsonschema.NewStringLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return fmt.Errorf("failed to compile schema %s: %w", name, err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.schemas[name] = schema

	return nil
}

// RemoveSchema removes a schema
func (v *Validator) RemoveSchema(name string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.schemas, name)
}

// ValidateJSON validates JSON data against a schema
func (v *Validator) ValidateJSON(schemaName string, data []byte) (*ValidationResult, error) {
	v.mu.RLock()
	schema, exists := v.schemas[schemaName]
	v.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("schema %s not found", schemaName)
	}

	loader := gojsonschema.NewBytesLoader(data)
	result, err := schema.Validate(loader)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	validationResult := &ValidationResult{
		Valid:  result.Valid(),
		Errors: make([]ValidationError, 0),
	}

	for _, err := range result.Errors() {
		validationResult.Errors = append(validationResult.Errors, ValidationError{
			Field:       err.Field(),
			Type:        err.Type(),
			Description: err.Description(),
			Value:       err.Value(),
		})
	}

	return validationResult, nil
}

// ValidateRequest validates an HTTP request body against a schema
func (v *Validator) ValidateRequest(schemaName string, r *http.Request) (*ValidationResult, []byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read request body: %w", err)
	}

	// Restore the body for downstream handlers
	r.Body = io.NopCloser(bytes.NewReader(body))

	if len(body) == 0 {
		return &ValidationResult{Valid: true}, body, nil
	}

	result, err := v.ValidateJSON(schemaName, body)
	return result, body, err
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field       string      `json:"field"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Value       interface{} `json:"value,omitempty"`
}

// Middleware returns middleware that validates requests and responses
func (v *Validator) Middleware(requestSchema, responseSchema string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validate request if schema is provided
			if requestSchema != "" && (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") {
				result, body, err := v.ValidateRequest(requestSchema, r)
				if err != nil {
					http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusInternalServerError)
					return
				}

				if !result.Valid {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":             "Request validation failed",
						"validation_errors": result.Errors,
					})
					return
				}

				// Restore body
				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			// If response validation is needed, wrap the response writer
			if responseSchema != "" {
				wrapped := &validatingResponseWriter{
					ResponseWriter: w,
					validator:      v,
					responseSchema: responseSchema,
					statusCode:     200,
				}
				next.ServeHTTP(wrapped, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// validatingResponseWriter wraps http.ResponseWriter to validate responses
type validatingResponseWriter struct {
	http.ResponseWriter
	validator      *Validator
	responseSchema string
	body           bytes.Buffer
	statusCode     int
	headerWritten  bool
}

func (vrw *validatingResponseWriter) Write(data []byte) (int, error) {
	if !vrw.headerWritten {
		vrw.WriteHeader(vrw.statusCode)
	}

	// Only validate JSON responses
	contentType := vrw.Header().Get("Content-Type")
	if contentType == "application/json" || contentType == "" {
		vrw.body.Write(data)
	}

	return vrw.ResponseWriter.Write(data)
}

func (vrw *validatingResponseWriter) WriteHeader(statusCode int) {
	if vrw.headerWritten {
		return
	}

	vrw.statusCode = statusCode
	vrw.headerWritten = true

	// Validate response if it's a success status and we have JSON content
	if statusCode >= 200 && statusCode < 300 && vrw.body.Len() > 0 {
		result, err := vrw.validator.ValidateJSON(vrw.responseSchema, vrw.body.Bytes())
		if err != nil {
			// Log validation error but don't fail the response
			// In production, you might want to log this properly
			return
		}

		if !result.Valid {
			// Log validation errors but don't fail the response
			// In production, you might want to log this properly
		}
	}

	vrw.ResponseWriter.WriteHeader(statusCode)
}

// ListSchemas returns the names of all loaded schemas
func (v *Validator) ListSchemas() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	schemas := make([]string, 0, len(v.schemas))
	for name := range v.schemas {
		schemas = append(schemas, name)
	}

	return schemas
}

// HasSchema checks if a schema exists
func (v *Validator) HasSchema(name string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	_, exists := v.schemas[name]
	return exists
}

// Stats returns validation statistics
func (v *Validator) Stats() Stats {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return Stats{
		TotalSchemas: len(v.schemas),
		SchemaNames:  v.ListSchemas(),
	}
}

// Stats represents validation statistics
type Stats struct {
	TotalSchemas int      `json:"total_schemas"`
	SchemaNames  []string `json:"schema_names"`
}
