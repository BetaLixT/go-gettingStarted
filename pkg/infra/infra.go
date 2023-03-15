package infra

import (
	"context"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"
	domtrace "techunicorn.com/udc-core/gettingStarted/pkg/domain/base/trace"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/config"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/lgr"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/psqldb"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/redisdb"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/snowflake"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/trace"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/trace/appinsights"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/trace/jaeger"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/trace/promex"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/tracelib"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/BetaLixT/gotred/v8"
	"github.com/BetaLixT/tsqlx"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	// Trace
	NewTraceExporterList,
	config.NewTraceOptions,
	trace.NewTracer,
	jaeger.NewJaegerTraceExporter,
	config.NewJaegerExporterOptions,
	appinsights.NewTraceExporter,
	config.NewAppInsightsExporterOptions,
	promex.NewTraceExporter,

	// Infra
	config.NewInitializer,
	lgr.NewLoggerFactory,
	wire.Bind(
		new(logger.IFactory),
		new(*lgr.LoggerFactory),
	),
	psqldb.NewDatabaseContext,
	wire.Bind(
		new(tsqlx.ITracer),
		new(*tracelib.Tracer),
	),
	config.NewPSQLDBOptions,
	redisdb.NewRedisContext,
	wire.Bind(
		new(gotred.ITracer),
		new(*tracelib.Tracer),
	),
	config.NewRedisOptions,
	snowflake.NewSnowflake,
	config.NewSnowflakeOptions,
	wire.Bind(
		new(domtrace.IRepository),
		new(*tracelib.Tracer),
	),
)

// NewTraceExporterList provides a list of exporters for tracing
func NewTraceExporterList(
	insexp appinsights.TraceExporter,
	jgrexp jaeger.TraceExporter,
	prmex promex.TraceExporter,
	lgrf logger.IFactory,
) *trace.ExporterList {
	lgr := lgrf.Create(context.Background())
	exp := []sdktrace.SpanExporter{}

	if insexp != nil {
		exp = append(exp, insexp)
	} else {
		lgr.Warn("insights exporter not found")
	}
	if jgrexp != nil {
		exp = append(exp, jgrexp)
	} else {
		lgr.Warn("jeager exporter not found")
	}
	if len(exp) == 0 {
		panic("no tracing exporters found (float you <3)")
	}
	exp = append(exp, prmex)
	return &trace.ExporterList{
		Exporters: exp,
	}
}
