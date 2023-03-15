package trace

import (
	"context"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/cntxt"
)

type traceExtractor struct{}

func (ex *traceExtractor) ExtractTraceInfo(
	c context.Context,
) (ver, tid, pid, rid, flg string) {
	ctx, ok := c.(cntxt.IContext)
	if !ok {
		ver = "00"
		tid = "0000000000000000"
		pid = "00000000"
		rid = "00000000"
		flg = "00"
		return
	}
	return ctx.GetTraceInfo()
}
