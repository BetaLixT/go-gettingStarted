// Package promex adds prometheus metrix
package promex

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// TraceExporter an application insights exporter
type TraceExporter sdktrace.SpanExporter

// NewTraceExporter constructs an app insights exporter
func NewTraceExporter() (TraceExporter, error) {
	return NewExporter("promex"), nil
}

type Exporter struct {
	requests      prometheus.Counter
	requestStatus prometheus.CounterVec
	responseTime  prometheus.HistogramVec
	bodySize      prometheus.HistogramVec

	events        prometheus.Counter
	eventsSuccess prometheus.CounterVec
	eventTime     prometheus.HistogramVec

	depStatus prometheus.CounterVec
}

// NewExporter constructs a new promex exporter
func NewExporter(prefix string) *Exporter {
	return &Exporter{
		requests: promauto.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_processed_reqs_total",
			Help: "The total number of processed requests",
		}),
		requestStatus: *promauto.NewCounterVec(prometheus.CounterOpts{
			Name: prefix + "_processed_reqs_status",
			Help: "The status codes of requests",
		}, []string{"code", "uri", "method", "ingress"}),
		responseTime: *promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    prefix + "_processed_reqs_latency",
			Help:    "The latency of requests",
			Buckets: []float64{0.0001, 0.0005, 0.0009, 0.001, 0.02, 0.05, 0.1, 0.3, 1.2, 5, 10},
		}, []string{"uri", "ingress"}),
		bodySize: *promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    prefix + "_processed_reqs_size",
			Help:    "The size of response bodies",
			Buckets: []float64{1, 1e+3, 50e+3, 100e+3, 250e+3, 500e+3, 750e+3, 1e+6, 250e+6, 500e+6, 750e+6, 1e+9, 10e+9},
		}, []string{"uri", "ingress"}),

		events: promauto.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_processed_evnts_total",
			Help: "The total number of processed events",
		}),
		eventsSuccess: *promauto.NewCounterVec(prometheus.CounterOpts{
			Name: prefix + "_processed_evnts_status",
			Help: "The status codes of events",
		}, []string{"status"}),
		eventTime: *promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: prefix + "_processed_evnts_latency",
			Help: "The latency of events",
		}, []string{"key"}),

		depStatus: *promauto.NewCounterVec(prometheus.CounterOpts{
			Name: prefix + "_dependencies_status",
			Help: "The status of dependencies",
		}, []string{"status", "type"}),
	}
}

func (exp *Exporter) Shutdown(ctx context.Context) error {
	return nil
}

// ExportSpans processes spans
func (exp *Exporter) ExportSpans(
	ctx context.Context,
	spans []sdktrace.ReadOnlySpan,
) error {
	for i := range spans {
		exp.process(spans[i])
	}
	return nil
}

// Preprocesses the Otel span and dispatches it to app insights differently
// based on the span kind.
func (exp *Exporter) process(sp sdktrace.ReadOnlySpan) {
	success := true
	if sp.Status().Code != codes.Ok {
		success = false
	}

	props := map[string]string{}

	switch sp.SpanKind() {
	case trace.SpanKindServer:
		exp.processRequest(sp)
	// TODO: i hate this
	case trace.SpanKindClient:
		rattr := sp.Resource().Attributes()
		for _, e := range rattr {
			props[string(e.Key)] = e.Value.AsString()
		}
		attr := sp.Attributes()
		for _, e := range attr {
			props[string(e.Key)] = e.Value.AsString()
		}
		exp.processDependency(sp, success, props)
	case trace.SpanKindProducer:
		rattr := sp.Resource().Attributes()
		for _, e := range rattr {
			props[string(e.Key)] = e.Value.AsString()
		}
		attr := sp.Attributes()
		for _, e := range attr {
			props[string(e.Key)] = e.Value.AsString()
		}
		exp.processDependency(sp, success, props)
	case trace.SpanKindConsumer:
		exp.processEvent()
	}
}

func (exp *Exporter) processRequest(
	sp sdktrace.ReadOnlySpan,
) {
	var url, responseCode, ingress, method string
	var bodySize int
	latency := sp.EndTime().Sub(sp.StartTime())
	found := 0
	rattr := sp.Resource().Attributes()
	for _, e := range rattr {
		if found == 4 {
			break
		}
		switch e.Key {
		case "url":
			url = e.Value.AsString()
			found++
		case "responseCode":
			responseCode = e.Value.AsString()
			found++
		case "ingress":
			ingress = e.Value.AsString()
			found++
		case "bodySize":
			bodySize = int(e.Value.AsInt64())
			found++
		case "method":
			method = e.Value.AsString()
			found++
		}
	}
	attr := sp.Attributes()
	for _, e := range attr {
		if found == 4 {
			break
		}
		switch e.Key {
		case "url":
			url = e.Value.AsString()
			found++
		case "responseCode":
			responseCode = e.Value.AsString()
			found++
		case "ingress":
			ingress = e.Value.AsString()
			found++
		case "bodySize":
			bodySize = int(e.Value.AsInt64())
			found++
		}
	}
	exp.requests.Inc()
	exp.requestStatus.WithLabelValues(responseCode, url, method, ingress).Inc()
	exp.responseTime.WithLabelValues(url, ingress).Observe(latency.Seconds())
	exp.bodySize.WithLabelValues(url, ingress).Observe(float64(bodySize))
}

func (exp *Exporter) processEvent() {
	exp.events.Inc()
	// TODO
}

func (exp *Exporter) processDependency(
	sp sdktrace.ReadOnlySpan,
	success bool,
	properties map[string]string,
) {
	typ := ""
	if val, ok := properties["type"]; ok {
		delete(properties, "type")
		typ = val
	}
	if success {
		exp.depStatus.WithLabelValues("success", typ).Inc()
	} else {
		exp.depStatus.WithLabelValues("failed", typ).Inc()
	}
}
