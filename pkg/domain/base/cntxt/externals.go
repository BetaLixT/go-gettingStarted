package cntxt

import (
	"context"
	"time"
)

// The context factory is to be used to create new contexts at the start of an
// incoming request
type IFactory interface {
	Create(
		traceparent string,
	) IContext
}

// An interface to the internally used context that only exposes functionality
// that is to be utilized in the domain layer
type IContext interface {
	context.Context
	CommitTransaction() error
	RollbackTransaction()
	SetTimeout(time.Duration)
	Cancel()
}
