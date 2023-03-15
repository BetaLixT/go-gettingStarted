package tracelib

import (
	"context"
	"time"

	"github.com/Soreing/motel"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Implement this to extract w3c-trace information from the context, an example
// usage would be to build a middleware for incoming http calls to inject trace
// information into the context and create an implementation of ITraceExtractor
// to get the trace (using ctx.Value and with the correct key(s))
type ITraceExtractor interface {
	ExtractTraceInfo(
		ctx context.Context,
	) (ver, tid, pid, rid, flg string)
}

type ISpanConstructor interface {
	NewRequestSpan(
		tid [16]byte,
		pid [8]byte,
		rid [8]byte,
		res *resource.Resource,
		method string,
		path string,
		query string,
		statusCode int,
		bodySize int,
		ip string,
		userAgent string,
		startTimestamp time.Time,
		eventTimestamp time.Time,
		fields map[string]string,
	) motel.Span
	NewEventSpan(
		tid [16]byte,
		pid [8]byte,
		rid [8]byte,
		res *resource.Resource,
		name string,
		key string,
		statusCode int,
		startTimestamp time.Time,
		eventTimestamp time.Time,
		fields map[string]string,
	) motel.Span
	NewDependencySpan(
		tid [16]byte,
		pid [8]byte,
		rid [8]byte,
		res *resource.Resource,
		dep *resource.Resource,
		dependencyType string,
		serviceName string,
		commandName string,
		success bool,
		startTimestamp time.Time,
		eventTimestamp time.Time,
		fields map[string]string,
	) motel.Span
}
