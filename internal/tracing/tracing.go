// Package tracing provides OpenTelemetry tracing initialization and configuration
// for the edge.link proxy service.
package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds tracing configuration
type Config struct {
	ServiceName    string `yaml:"service_name" json:"service_name"`
	ServiceVersion string `yaml:"service_version" json:"service_version"`
	Environment    string `yaml:"environment" json:"environment"`
	OTLPEndpoint   string `yaml:"otlp_endpoint" json:"otlp_endpoint"`
	Enabled        bool   `yaml:"enabled" json:"enabled"`
}

// Provider wraps the OpenTelemetry trace provider
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
}

// NewProvider creates a new OpenTelemetry tracing provider
func NewProvider(cfg Config) (*Provider, error) {
	if !cfg.Enabled {
		// Return a no-op provider when tracing is disabled
		return &Provider{}, nil
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // Use insecure for local development
	)
	if err != nil {
		return nil, err
	}

	// Create trace provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer for this service
	tracer := tracerProvider.Tracer(cfg.ServiceName)

	return &Provider{
		tracerProvider: tracerProvider,
		tracer:         tracer,
	}, nil
}

// Tracer returns the OpenTelemetry tracer
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// Shutdown gracefully shuts down the trace provider
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tracerProvider == nil {
		return nil
	}
	return p.tracerProvider.Shutdown(ctx)
}

// StartSpan starts a new trace span with common attributes
func (p *Provider) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if p.tracer == nil {
		// Return no-op span when tracing is disabled
		return ctx, trace.SpanFromContext(ctx)
	}
	return p.tracer.Start(ctx, spanName, opts...)
}

// AddCommonAttributes adds common attributes to a span
func AddCommonAttributes(span trace.Span, tenantID int, routeID int, requestID, apiKeyPrefix string) {
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.Int("tenant.id", tenantID),
	}

	if routeID > 0 {
		attrs = append(attrs, attribute.Int("route.id", routeID))
	}

	if requestID != "" {
		attrs = append(attrs, attribute.String("request.id", requestID))
	}

	if apiKeyPrefix != "" {
		attrs = append(attrs, attribute.String("api_key.prefix", apiKeyPrefix))
	}

	span.SetAttributes(attrs...)
}

// SpanFromContext returns the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// DefaultConfig returns a default tracing configuration
func DefaultConfig() Config {
	return Config{
		ServiceName:    "edgelink-proxy",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		OTLPEndpoint:   "http://localhost:4318", // Default OTLP HTTP endpoint
		Enabled:        true,
	}
}
