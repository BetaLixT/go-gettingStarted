// Package appinsights provides an app insights exporter for tracing
package appinsights

import (
	"github.com/Soreing/apex"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TraceExporter an application insights exporter
type TraceExporter sdktrace.SpanExporter

// NewTraceExporter constructs an app insights exporter
func NewTraceExporter(
	opts *ExporterOptions,
) (TraceExporter, error) {
	if opts.InstrKey == "" {
		return nil, nil
	}

	return apex.NewExporter(
		opts.InstrKey,
		func(msg string) error {
			return nil
		},
	)
}
