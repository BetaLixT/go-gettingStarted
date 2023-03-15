package logger

import (
	"context"

	"go.uber.org/zap"
)

// Logger factory to be used to create a new logger
type IFactory interface {
	Create(ctx context.Context) *zap.Logger
}
