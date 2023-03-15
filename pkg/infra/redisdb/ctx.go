// Package redisdb contains constructors, options and sets up the tracing plugin
// for the redis client library
package redisdb

import (
	"context"
	"crypto/tls"

	"github.com/BetaLixT/gotred/v8"
	"github.com/go-redis/redis/v8"
)

// NewRedisContext constructs a new redis database client
func NewRedisContext(
	optn *Options,
	tracer gotred.ITracer,
) (*redis.Client, error) {
	rop := &redis.Options{
		Addr:     optn.Address,
		Password: optn.Password, // no password set
		DB:       optn.Database, // use default DB
	}
	if optn.TLS {
		rop.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			// Certificates: []tls.Certificate{cert}
		}
	}
	client := redis.NewClient(
		rop,
	)
	ctx := context.Background()
	status := client.Ping(ctx)
	err := status.Err()
	if err != nil {
		return nil, err
	}
	traceHook := gotred.NewTraceHook(
		tracer,
		optn.ServiceName,
	)

	client.AddHook(traceHook)
	return client, nil
}
