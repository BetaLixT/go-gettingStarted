package repos

import (
	"context"
	"fmt"
	"sync"
	"time"

	domcntxt "techunicorn.com/udc-core/gettingStarted/pkg/domain/base/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"
	implcntxt "techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/cntxt"
	infrcntxt "techunicorn.com/udc-core/gettingStarted/pkg/infra/cntxt"

	"github.com/BetaLixT/go-resiliency/retrier"
	"go.uber.org/zap"
)

// =============================================================================
// Internal context implementation
// =============================================================================

var (
	_ context.Context    = (*internalContext)(nil)
	_ domcntxt.IContext  = (*internalContext)(nil)
	_ infrcntxt.IContext = (*internalContext)(nil)
	_ implcntxt.IContext = (*internalContext)(nil)
)

type internalContext struct {
	lgrf logger.IFactory
	// deadline time.Time
	cancelmtx *sync.Mutex
	err       error
	done      chan struct{}
	dur       time.Time

	// - transaction
	rtr                 retrier.Retrier
	compensatoryActions []implcntxt.Action
	commitActions       []implcntxt.Action
	events              []dispatchableEvent
	txObjs              map[string]interface{}
	isCommited          bool
	isRolledback        bool
	txmtx               *sync.Mutex

	// trace
	ver string
	tid string
	pid string
	rid string
	flg string
}

// - Base context functions
func (c *internalContext) cancel(err error) {
	c.RollbackTransaction()
	c.cancelmtx.Lock()
	defer c.cancelmtx.Unlock()
	if c.err != nil {
		return
	}
	c.err = err
	close(c.done)
}

func (c *internalContext) Cancel() {
	c.cancel(fmt.Errorf("context manually canceled"))
}

func (c *internalContext) Deadline() (time.Time, bool) {
	return time.Now(), false
}

func (c *internalContext) Done() <-chan struct{} {
	return c.done
}

func (c *internalContext) Err() error {
	return c.err
}

func (c *internalContext) Value(key any) any {
	return nil
}

// - Transaction functions
func (c *internalContext) SetTimeout(timeout time.Duration) {
	c.dur = time.Now().Add(timeout)
	time.AfterFunc(
		time.Until(c.dur),
		func() {
			c.cancel(context.DeadlineExceeded)
		},
	)
}

func (c *internalContext) CommitTransaction() error {
	c.txmtx.Lock()
	defer c.txmtx.Unlock()
	if c.isCommited || c.isRolledback {
		return fmt.Errorf(
			"tried to commit transaction that has already been commited/rolled back",
		)
	}
	ctx := newMinimalContext(c)
	for range c.events {
		// TODO Event handling
		// err := c.ndisp.DispatchEventNotification(
		// 	ctx,
		// 	evnt.stream,
		// 	evnt.streamId,
		// 	evnt.event,
		// 	evnt.version,
		// 	evnt.data,
		// 	evnt.eventTime,
		// )
		// if err != nil {
		// 	return err
		// }
	}
	for _, commit := range c.commitActions {
		err := commit(ctx)
		if err != nil {
			return err
		}
	}
	// TODO: confirm event
	c.isCommited = true
	return nil
}

// TODO: better handling failed rollback transaction
func (c *internalContext) RollbackTransaction() {
	c.txmtx.Lock()
	defer c.txmtx.Unlock()
	if c.isCommited || c.isRolledback {
		return
	}

	c.isRolledback = true
	ctx := newMinimalContext(c)
	lgr := c.lgrf.Create(ctx)
	for _, cmp := range c.compensatoryActions {
		err := c.rtr.Run(func() error {
			err := cmp(ctx)
			if err != nil {
				lgr.Warn("failed to run compensatory action", zap.Error(err))
			}
			return err
		})
		if err != nil {
			lgr.Error(
				"failed to run compensatory action, max retries exceeded",
				zap.Error(err),
			)
		}
	}
}

func (c *internalContext) RegisterCompensatoryAction(
	cmp ...implcntxt.Action,
) {
	c.txmtx.Lock()
	defer c.txmtx.Unlock()
	c.compensatoryActions = append(c.compensatoryActions, cmp...)
}

func (c *internalContext) RegisterCommitAction(
	cmp ...implcntxt.Action,
) {
	c.txmtx.Lock()
	defer c.txmtx.Unlock()
	c.commitActions = append(c.commitActions, cmp...)
}

func (c *internalContext) RegisterEvent(
	id uint64,
	sagaID *string,
	stream string,
	streamID string,
	event string,
	version uint64,
	eventTime time.Time,
	data interface{},
) {
	c.txmtx.Lock()
	defer c.txmtx.Unlock()
	c.events = append(c.events, dispatchableEvent{
		stream:    stream,
		streamID:  streamID,
		event:     event,
		version:   int(version),
		eventTime: eventTime,
		data:      data,
	})
}

func (c *internalContext) GetTransactionObject(
	key string,
	constr implcntxt.Constructor,
) (interface{}, bool, error) {
	c.txmtx.Lock()
	defer c.txmtx.Unlock()
	intr, ok := c.txObjs[key]
	if ok {
		return intr, false, nil
	}
	intr, err := constr()
	if err != nil {
		return nil, false, err
	}
	c.txObjs[key] = intr
	return intr, true, nil
}

func (c *internalContext) GetTraceInfo() (ver, tid, pid, rid, flg string) {
	return c.ver, c.tid, c.pid, c.rid, c.flg
}

// - Minimal context
