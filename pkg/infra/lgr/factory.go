package lgr

import (
	"context"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/cntxt"

	"go.uber.org/zap"
)

type LoggerFactory struct {
	lgr *zap.Logger
}

// func (f*LoggerFactory) Create(ctx context.Context) *zap.Logger

func NewLoggerFactory() (*LoggerFactory, error) {
	lgr, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &LoggerFactory{
		lgr: lgr,
	}, nil
}

func (lf *LoggerFactory) Create(
	c context.Context,
) *zap.Logger {
	if c == nil {
		return lf.lgr
	}
	ctx, ok := c.(cntxt.IContext)
	if !ok {
		return lf.lgr
	}

	_, tid, pid, rid, _ := ctx.GetTraceInfo()
	return lf.lgr.With(
		zap.String("tid", tid),
		zap.String("pid", pid),
		zap.String("rid", rid),
	)
}

func (lf *LoggerFactory) Close() {
	lf.lgr.Sync()
}
