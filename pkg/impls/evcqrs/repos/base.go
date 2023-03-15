// Package repos implements the interfaces defined on the domain layer
package repos

import (
	"context"

	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/common"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/entities"

	"github.com/BetaLixT/tsqlx"
)

// BaseDataRepository is the base repository containing common database
// functionality to be embeded by other repos that implement persistence to the
// database
type BaseDataRepository struct {
	dbctx *tsqlx.TracedDB
}

// NewBaseDataRepository Constructs a new base data repository
func NewBaseDataRepository(
	dbctx *tsqlx.TracedDB,
) *BaseDataRepository {
	return &BaseDataRepository{
		dbctx: dbctx,
	}
}

func (r *BaseDataRepository) insertEvent(
	ctx cntxt.IContext,
	trctx *tsqlx.TracedTx,
	out entities.IBaseEvent,
	sagaID *string,
	stream string,
	id string,
	version uint64,
	event string,
	data interface{},
) error {
	_, tid, _, rid, _ := ctx.GetTraceInfo()
	err := trctx.Get(
		ctx,
		out,
		insertEventQuery,
		sagaID,
		stream,
		id,
		version,
		event,
		tid,
		rid,
		data,
	)
	if err == nil {
		ctx.RegisterEvent(
			out.GetID(),
			sagaID,
			stream,
			id,
			event,
			version,
			out.GetEventTime(),
			data,
		)
	}
	return err
}

func (r *BaseDataRepository) getDBTx(
	ctx cntxt.IContext,
) (*tsqlx.TracedTx, error) {
	idbtx, nw, err := ctx.GetTransactionObject(
		common.SqlTransactionObjectKey,
		func() (interface{}, error) {
			return r.dbctx.Beginx()
		},
	)
	if err != nil {
		return nil, err
	}

	dbtx, ok := idbtx.(*tsqlx.TracedTx)
	if !ok {
		return nil, common.NewFailedToAssertDatabaseCtxTypeError()
	}
	if nw {
		ctx.RegisterCommitAction(func(ctx context.Context) error {
			return dbtx.Commit()
		})
		ctx.RegisterCompensatoryAction(func(ctx context.Context) error {
			return dbtx.Rollback()
		})
	}
	return dbtx, nil
}

// GetValueOrDefault either returns the value if the provided pointer is not nil
// else it provides the default value
func GetValueOrDefault[v comparable](in *v) (out v) {
	if in != nil {
		out = *in
	}
	return out
}

const (
	insertEventQuery = `
	INSERT INTO events(
		saga_id,
		stream,
		stream_id,
		version,
		event,
		trace_id,
		request_id,
		data
	) VALUES(
		$1, $2, $3, $4, $5, $6, $7, $8
	) RETURNING *`
)
