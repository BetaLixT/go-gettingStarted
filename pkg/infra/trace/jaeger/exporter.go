// Package jaeger defines the jaeger exporter for tracing
package jaeger

import (
	jexp "go.opentelemetry.io/otel/exporters/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TraceExporter exporter for jaeger
type TraceExporter sdktrace.SpanExporter

// NewJaegerTraceExporter Constrcuts a new jeager trace exporter
func NewJaegerTraceExporter(
	opts *ExporterOptions,
) (TraceExporter, error) {
	if opts.Endpoint == "" {
		return nil, nil
	}
	return jexp.New(
		jexp.WithCollectorEndpoint(
			jexp.WithEndpoint(opts.Endpoint),
		),
	)
}
