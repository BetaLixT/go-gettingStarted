package repos

import (
	"context"
	"time"

	infrcntxt "techunicorn.com/udc-core/gettingStarted/pkg/infra/cntxt"
)

var (
	_ context.Context    = (*minimalContext)(nil)
	_ infrcntxt.IContext = (*minimalContext)(nil)
)

func newMinimalContext(ctx *internalContext) *minimalContext {
	return &minimalContext{
		done: make(chan struct{}, 1),
		ver:  ctx.ver,
		tid:  ctx.tid,
		pid:  ctx.pid,
		rid:  ctx.rid,
		flg:  ctx.flg,
	}
}

type minimalContext struct {
	done chan struct{}
	ver  string
	tid  string
	pid  string
	rid  string
	flg  string
}

// - Base context functions
func (c *minimalContext) Deadline() (time.Time, bool) {
	return time.Now(), false
}

func (c *minimalContext) Done() <-chan struct{} {
	return c.done
}

func (c *minimalContext) Err() error {
	return nil
}

func (c *minimalContext) Value(key any) any {
	return nil
}

func (c *minimalContext) GetTraceInfo() (ver, tid, pid, rid, flg string) {
	return c.ver, c.tid, c.pid, c.rid, c.flg
}

type dispatchableEvent struct {
	stream    string
	streamID  string
	sagaID    *string
	version   int
	event     string
	eventTime time.Time
	data      interface{}
}
