package cntxt

import (
	"context"
	"time"
)

type Action func(context.Context) error
type Constructor func() (interface{}, error)

type IContext interface {
	context.Context
	RegisterCompensatoryAction(...Action)
	RegisterCommitAction(...Action)
	RegisterEvent(
		id uint64,
		sagaId *string,
		stream string,
		streamId string,
		event string,
		version uint64,
		eventTime time.Time,
		data interface{},
	)
	GetTransactionObject(
		key string,
		constr Constructor,
	) (obj interface{}, isnew bool, err error)
	GetTraceInfo() (ver, tid, pid, rid, flg string)
}
