package tracelib

import (
	"context"
	"fmt"
	"time"

	"github.com/Soreing/motel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

type DefaultTraceExtractor struct{}

func (*DefaultTraceExtractor) ExtractTraceInfo(
	_ context.Context,
) (ver, tid, pid, rid, flg string) {
	return "", "", "", "", ""
}

type DefaultSpanConstructor struct{}

// NewRequestSpan create a new request span
func (sc *DefaultSpanConstructor) NewRequestSpan(
	tid [16]byte, pid [8]byte, rid [8]byte,
	res *resource.Resource,
	method string, path string, query string,
	statusCode int, bodySize int, ip string,
	userAgent string,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) motel.Span {
	span := motel.CreateSpan(
		fmt.Sprintf("%s %s", method, path),
		trace.SpanKindServer,
		res, tid, pid, rid, 0x01,
		statusCode > 99 && statusCode < 300,
		startTimestamp, eventTimestamp,
	)
	return span
}

// NewEventSpan create new event span
func (sc *DefaultSpanConstructor) NewEventSpan(
	tid [16]byte, pid [8]byte, rid [8]byte,
	res *resource.Resource,
	name string, key string, statusCode int,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) motel.Span {
	span := motel.CreateSpan(
		name,
		trace.SpanKindConsumer,
		res, tid, pid, rid, 0x01,
		statusCode > 99 && statusCode < 300,
		startTimestamp, eventTimestamp,
	)
	return span
}

// NewDependencySpan creates a new dependency span
func (sc *DefaultSpanConstructor) NewDependencySpan(
	tid [16]byte, pid [8]byte, rid [8]byte,
	res *resource.Resource,
	dep *resource.Resource,
	dependencyType string,
	serviceName string,
	commandName string,
	success bool,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) motel.Span {
	span := motel.CreateSpan(
		commandName,
		trace.SpanKindClient,
		res, tid, pid, rid, 0x01,
		success, startTimestamp, eventTimestamp,
	)
	return span
}
