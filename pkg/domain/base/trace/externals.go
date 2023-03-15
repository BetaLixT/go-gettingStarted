package trace

import (
	"context"
	"time"
)

type IRepository interface {
	TraceRequest(
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
	)
}
