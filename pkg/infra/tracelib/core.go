// Package tracelib defines immplementation for the ITracer interfaces used by
// our libraries
package tracelib

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	mrand "math/rand"
	"time"

	"github.com/Soreing/motel"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"go.uber.org/zap"
)

// Tracer implementation of ITracer
type Tracer struct {
	constructor ISpanConstructor
	extractor   ITraceExtractor
	collector   motel.SpanCollector
	exporters   []sdktrace.SpanExporter
	resource    *resource.Resource
	rand        *mrand.Rand
}

// Creates a random number generator
func makeRand() (rng *mrand.Rand, err error) {
	bseed := make([]byte, 8)
	n, err := crand.Read(bseed)
	if err != nil || n != 8 {
		return nil, errors.New("failed to read seed")
	}

	return mrand.New(
		mrand.NewSource(
			int64(binary.BigEndian.Uint64(bseed)),
		),
	), nil
}

// NewBasic Constructs an instance of Tracer with defaults, including
// the default ITraceExtractor which does not provide any tracing information
// from the context it is recommended to use the non context dependent functions
// (functions that end with "WithIds") to take advangate of the tracing if you
// use this constructor
func NewBasic(
	serviceName string,
	exporters []sdktrace.SpanExporter,
) (*Tracer, error) {
	rand, err := makeRand()
	if err != nil {
		return nil, errors.New("failed to create random")
	}

	res, err := resource.New(
		context.TODO(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, errors.New("failed to create resource")
	}

	sc := motel.NewSpanCollector(exporters, 0, 0)
	return &Tracer{
		constructor: &DefaultSpanConstructor{},
		collector:   sc,
		exporters:   exporters,
		resource:    res,
		rand:        rand,
	}, nil
}

// NewBasicWithLogger Constructs an instance of Tracer using the
// provided zap logger and the default ITraceExtractor which does not provide
// any tracing information from the context it is recommended to use the non
// context dependent functions (functions that end with "WithIds") to take
// advangate of the tracing if you use this constructor
func NewBasicWithLogger(
	serviceName string,
	exporters []sdktrace.SpanExporter,
	lgr zap.Logger,
) (*Tracer, error) {
	rand, err := makeRand()
	if err != nil {
		return nil, errors.New("failed to create random")
	}

	res, err := resource.New(
		context.TODO(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, errors.New("failed to create resource")
	}

	sc := motel.NewSpanCollector(exporters, 0, 0)
	return &Tracer{
		constructor: &DefaultSpanConstructor{},
		collector:   sc,
		exporters:   exporters,
		resource:    res,
		rand:        rand,
	}, nil
}

// NewTracer Constructs an instance of Tracer using the provided zap
// logger and a custom trace extractor, it's recommended to provide a custom
// trace extractor that will extract the w3c trace information from the context
// and take advantage of the context dependent trace functions, check
// documentation of ITraceExtractor for more information
func NewTracer(
	serviceName string,
	exporters []sdktrace.SpanExporter,
	constructor ISpanConstructor,
	extractor ITraceExtractor,
	lgr *zap.Logger,
) (*Tracer, error) {
	rand, err := makeRand()
	if err != nil {
		return nil, errors.New("failed to create random")
	}

	res, err := resource.New(
		context.TODO(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, errors.New("failed to create resource")
	}

	sc := motel.NewSpanCollector(exporters, -1, 0)
	return &Tracer{
		collector:   sc,
		constructor: constructor,
		extractor:   extractor,
		exporters:   exporters,
		resource:    res,
		rand:        rand,
	}, nil
}

// Close does nothing, just satisfying interfaces
func (ins *Tracer) Close() {
}

// Converts string Trace Ids to binary values in byte arrays
func stobTraceIds(
	tidStr string,
	pidStr string,
	sidStr string,
) (tid [16]byte, pid [8]byte, sid [8]byte) {
	tidSlc, err := hex.DecodeString(tidStr)
	if err != nil {
		tidSlc = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}
	pidSlc, err := hex.DecodeString(pidStr)
	if err != nil {
		pidSlc = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	}
	sidSlc, err := hex.DecodeString(sidStr)
	if err != nil {
		sidSlc = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	}

	copy(tid[:], tidSlc)
	copy(pid[:], pidSlc)
	copy(sid[:], sidSlc)
	return
}

// CreateResourceIdBytes Creates new resource id as a byte array
func (ins *Tracer) CreateResourceIdBytes() (rid [8]byte, err error) {
	ridSlc, n := make([]byte, 8), 0
	if n, err = ins.rand.Read(ridSlc); err != nil || n != 8 {
		return rid, err
	}
	copy(rid[:], ridSlc)
	return
}

// CreateResourceIdString Creates new resource id as a string
func (ins *Tracer) CreateResourceIdString() (rid string, err error) {
	ridSlc, n := make([]byte, 8), 0
	if n, err = ins.rand.Read(ridSlc); err != nil || n != 8 {
		return rid, err
	}
	return hex.EncodeToString(ridSlc), nil
}

// ExtractTraceInfo !! - This only needed by older interfaces
func (ins *Tracer) ExtractTraceInfo(
	ctx context.Context,
) (ver, tid, pid, rid, flg string) {
	return ins.extractor.ExtractTraceInfo(ctx)
} // !! - End of legacy function c:

// - Context dependent

// TraceRequest will trace an incoming request
func (ins *Tracer) TraceRequest(
	ctx context.Context,
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
) {
	_, tid, pid, rid, _ := ins.extractor.ExtractTraceInfo(ctx)
	tidb, pidb, ridb := stobTraceIds(tid, pid, rid)

	span := ins.constructor.NewRequestSpan(
		tidb, pidb, ridb, ins.resource,
		method, path, query, statusCode, bodySize, ip,
		userAgent, startTimestamp, eventTimestamp, fields,
	)
	ins.collector.Feed(span)
}

// TraceEvent will trace an incomming event
func (ins *Tracer) TraceEvent(
	ctx context.Context,
	name string,
	key string,
	statusCode int,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {
	_, tid, pid, rid, _ := ins.extractor.ExtractTraceInfo(ctx)
	tidb, pidb, ridb := stobTraceIds(tid, pid, rid)

	span := ins.constructor.NewEventSpan(
		tidb, pidb, ridb, ins.resource,
		name, key, statusCode, startTimestamp, eventTimestamp, fields,
	)
	ins.collector.Feed(span)
}

// TraceDependency will trace calls to external dependencies
func (ins *Tracer) TraceDependency(
	ctx context.Context,
	spanId string,
	dependencyType string,
	serviceName string,
	commandName string,
	success bool,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {
	_, tid, _, rid, _ := ins.extractor.ExtractTraceInfo(ctx)
	tidb, pidb, sidb := stobTraceIds(tid, rid, spanId)

	res, _ := resource.New(
		context.TODO(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	span := ins.constructor.NewDependencySpan(
		tidb, pidb, sidb, ins.resource, res,
		dependencyType, serviceName, commandName,
		success, startTimestamp, eventTimestamp, fields,
	)
	ins.collector.Feed(span)
}

// - Context Independent

// TraceRequestWithIds trace incoming requests without a context
func (ins *Tracer) TraceRequestWithIds(
	traceId string,
	parentId string,
	requestId string,
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
) {
	tidb, pidb, ridb := stobTraceIds(traceId, parentId, requestId)

	span := ins.constructor.NewRequestSpan(
		tidb, pidb, ridb, ins.resource,
		method, path, query, statusCode, bodySize, ip,
		userAgent, startTimestamp, eventTimestamp, fields,
	)
	ins.collector.Feed(span)
}

// TraceEventWithIds trace incoming events without a context
func (ins *Tracer) TraceEventWithIds(
	traceId string,
	parentId string,
	requestId string,
	name string,
	key string,
	statusCode int,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {
	tidb, pidb, ridb := stobTraceIds(traceId, parentId, requestId)

	span := ins.constructor.NewEventSpan(
		tidb, pidb, ridb, ins.resource,
		name, key, statusCode, startTimestamp, eventTimestamp, fields,
	)
	ins.collector.Feed(span)
}

// TraceDependencyWithIds trace dependencies without a context
func (ins *Tracer) TraceDependencyWithIds(
	traceId string,
	requestId string,
	spanId string,
	dependencyType string,
	serviceName string,
	commandName string,
	success bool,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {
	tidb, pidb, sidb := stobTraceIds(traceId, requestId, spanId)

	res, _ := resource.New(
		context.TODO(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	span := ins.constructor.NewDependencySpan(
		tidb, pidb, sidb, ins.resource, res,
		dependencyType, serviceName, commandName,
		success, startTimestamp, eventTimestamp, fields,
	)
	ins.collector.Feed(span)
}
