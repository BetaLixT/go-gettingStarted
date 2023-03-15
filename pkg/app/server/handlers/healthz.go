package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/betalixt/gorr"
	"go.uber.org/zap"
	"techunicorn.com/udc-core/gettingStarted/pkg/app/server/common"
	srvcontracts "techunicorn.com/udc-core/gettingStarted/pkg/app/server/contracts"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/contracts"
)

type HealthzHandler struct {
	srvcontracts.UnimplementedHealthzServer
	ctxf cntxt.IFactory
	lgrf logger.IFactory
}

var _ srvcontracts.HealthzServer = (*HealthzHandler)(nil)

func NewHealthzHandler(
	ctxf cntxt.IFactory,
	lgrf logger.IFactory,
) *HealthzHandler {
	return &HealthzHandler{
		ctxf: ctxf,
		lgrf: lgrf,
	}
}

func (h *HealthzHandler) GetHealthStatus(
	c context.Context,
	qry *contracts.HealthQuery,
) (res *contracts.EmptyResponse, err error) {
	if qry.UserContext == nil {
		return nil, common.NewUserContextMissingError()
	}
	ctx, ok := c.(cntxt.IContext)
	if !ok {
		return nil, common.NewInvalidContextProvidedToHandlerError()
	}
	ctx.SetTimeout(2 * time.Minute)
	lgr := h.lgrf.Create(ctx)
	lgr.Info(
		"handling",
		zap.Any("qry", qry),
	)
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = gorr.NewUnexpectedError(fmt.Errorf("%v", r))
				lgr.Error(
					"root panic recovered handling request",
					zap.Any("panic", r),
					zap.Stack("stack"),
				)
			} else {
				lgr.Error(
					"root panic recovered handling request",
					zap.Error(err),
					zap.Stack("stack"),
				)
			}
			ctx.RollbackTransaction()
			ctx.Cancel()
		}
		if err != nil {
			if _, ok := err.(*gorr.Error); !ok {
				err = gorr.NewUnexpectedError(err)
			}
		}
		return
	}()
	res = &contracts.EmptyResponse{}
	if err != nil {
		lgr.Error(
			"command handling failed",
			zap.Error(err),
		)
		ctx.RollbackTransaction()
	} else {
		err = ctx.CommitTransaction()
		if err != nil {
			lgr.Error(
				"failed to commit transaction",
				zap.Error(err),
			)
			ctx.RollbackTransaction()
		}
	}
	ctx.Cancel()
	return
}
