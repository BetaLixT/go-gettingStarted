package trace

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Soreing/motel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

type spanConstructor struct{}

func (sc *spanConstructor) NewRequestSpan(
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
) motel.Span {
	span := motel.CreateSpan(
		fmt.Sprintf("%s %s", method, path),
		trace.SpanKindServer,
		res, tid, pid, rid, 0x01,
		statusCode > 99 && statusCode < 300,
		startTimestamp, eventTimestamp,
	)
	span.WithAttribute(
		"responseCode",
		attribute.StringValue(strconv.Itoa(statusCode)),
	)
	if ingress, ok := fields["ingress"]; ok {
		span.WithAttribute("ingress", attribute.StringValue(ingress))
	}

	span.WithAttribute("method", attribute.StringValue(method))
	span.WithAttribute("url", attribute.StringValue(path+query))
	span.WithAttribute("bodySize", attribute.IntValue(bodySize))
	span.WithAttribute("ip", attribute.StringValue(ip))
	span.WithAttribute("userAgent", attribute.StringValue(userAgent))
	return span
}

func (sc *spanConstructor) NewEventSpan(
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
) motel.Span {
	span := motel.CreateSpan(
		name,
		trace.SpanKindConsumer,
		res, tid, pid, rid, 0x01,
		statusCode > 99 && statusCode < 300,
		startTimestamp, eventTimestamp,
	)
	span.WithAttribute(
		"responseCode",
		attribute.StringValue(strconv.Itoa(statusCode)),
	)
	span.WithAttribute("key", attribute.StringValue(key))
	return span
}

func (sc *spanConstructor) NewDependencySpan(
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
) motel.Span {
	span := motel.CreateSpan(
		commandName,
		trace.SpanKindClient,
		dep, tid, pid, rid, 0x01,
		success, startTimestamp, eventTimestamp,
	)

	for _, e := range res.Attributes() {
		if string(e.Key) == string(semconv.ServiceNameKey) {
			span.WithAttribute("source", e.Value)
		}
	}

	span.WithAttribute("type", attribute.StringValue(dependencyType))
	return span
}
