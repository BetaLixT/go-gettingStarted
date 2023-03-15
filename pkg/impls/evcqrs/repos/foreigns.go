package repos

import (
	"context"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/foreigns"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/common"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/entities"

	"go.uber.org/zap"
)

type ForeignsRepository struct {
	*BaseDataRepository
	lgrf logger.IFactory
}

var _ foreigns.IRepository = (*ForeignsRepository)(nil)

func NewForeignsRepository(
	base *BaseDataRepository,
	lgrf logger.IFactory,
) *ForeignsRepository {
	return &ForeignsRepository{
		BaseDataRepository: base,
		lgrf:               lgrf,
	}
}

func (r *ForeignsRepository) RegisterForeignItem(
	c context.Context,
	sagaId *string,
	foreignStream string,
	foreignStreamId string,
) error {
	lgr := r.lgrf.Create(c)

	ctx, ok := c.(cntxt.IContext)
	if !ok {
		lgr.Error("unexpected context type")
		return common.NewFailedToAssertContextTypeError()
	}

	dbtx, err := r.getDBTx(ctx)
	if err != nil {
		lgr.Error("failed to get db transaction", zap.Error(err))
		return err
	}

	entry := entities.Foreign{}
	err = dbtx.Get(
		ctx,
		&entry,
		InsertForeignItemQuery,
		foreignStream,
		foreignStreamId,
		sagaId,
	)
	if err != nil {
		lgr.Error("failed to insert foreign item", zap.Error(err))
	}
	return err
}

func (r *ForeignsRepository) RemoveForeignItem(
	c context.Context,
	foreignStream string,
	foreignStreamId string,
) error {
	lgr := r.lgrf.Create(c)

	ctx, ok := c.(cntxt.IContext)
	if !ok {
		lgr.Error("unexpected context type")
		return common.NewFailedToAssertContextTypeError()
	}

	dbtx, err := r.getDBTx(ctx)
	if err != nil {
		lgr.Error("failed to get db transaction", zap.Error(err))
		return err
	}

	entry := entities.Foreign{}
	err = dbtx.Get(
		ctx,
		&entry,
		DeleteForeignItemQuery,
		foreignStream,
		foreignStreamId,
	)
	if err != nil {
		lgr.Error("failed to insert foreign item", zap.Error(err))
	}
	return err
}

func (r *ForeignsRepository) RegisterConstraint(
	c context.Context,
	sagaId *string,
	foreignStream string,
	foreignStreamId string,
	stream string,
	streamId string,
) error {
	lgr := r.lgrf.Create(c)

	ctx, ok := c.(cntxt.IContext)
	if !ok {
		lgr.Error("unexpected context type")
		return common.NewFailedToAssertContextTypeError()
	}

	dbtx, err := r.getDBTx(ctx)
	if err != nil {
		lgr.Error("failed to get db transaction", zap.Error(err))
		return err
	}

	entry := entities.ForeignConstraint{}
	err = dbtx.Get(
		ctx,
		&entry,
		InsertForeignConstraintQuery,
		foreignStream,
		foreignStreamId,
		stream,
		streamId,
		sagaId,
	)
	if err != nil {
		lgr.Error("failed to insert foreign constraint", zap.Error(err))
	}
	return err
}

func (r *ForeignsRepository) RemoveConstraint(
	c context.Context,
	foreignStream string,
	foreignStreamId string,
	stream string,
	streamId string,
) error {
	lgr := r.lgrf.Create(c)

	ctx, ok := c.(cntxt.IContext)
	if !ok {
		lgr.Error("unexpected context type")
		return common.NewFailedToAssertContextTypeError()
	}

	dbtx, err := r.getDBTx(ctx)
	if err != nil {
		lgr.Error("failed to get db transaction", zap.Error(err))
		return err
	}

	entry := entities.ForeignConstraint{}
	err = dbtx.Get(
		ctx,
		&entry,
		DeleteForeignConstraintQuery,
		foreignStream,
		foreignStreamId,
		stream,
		streamId,
	)
	if err != nil {
		lgr.Error("failed to remove foreign constraint", zap.Error(err))
	}
	return err
}

func (r *ForeignsRepository) ListAssociatedObjects(
	c context.Context,
	foreignStream string,
	foreignStreamId string,
) ([]foreigns.Object, error) {
	lgr := r.lgrf.Create(c)

	ctx, ok := c.(cntxt.IContext)
	if !ok {
		lgr.Error("unexpected context type")
		return nil, common.NewFailedToAssertContextTypeError()
	}

	dbtx, err := r.getDBTx(ctx)
	if err != nil {
		lgr.Error("failed to get db transaction", zap.Error(err))
		return nil, err
	}

	entries := []entities.ForeignAssociatedObject{}
	err = dbtx.Select(
		ctx,
		&entries,
		ListAssociatedObjectsQuery,
		foreignStream,
		foreignStreamId,
	)
	if err != nil {
		lgr.Error("failed to remove foreign constraint", zap.Error(err))
		return nil, err
	}
	return ((*entities.ForeignAssociatedObject)(nil)).ToDTOSlice(
		entries,
	), nil
}

const (
	InsertForeignItemQuery = `
	INSERT INTO foreigns(
		stream,
		stream_id,
		saga_id
	) VALUES(
		$1, $2, $3
	) RETURNING *
	`

	DeleteForeignItemQuery = `
	DELETE FROM foreigns
	WHERE stream = $1 AND stream_id = $2
	RETURNING *
	`

	InsertForeignConstraintQuery = `
	INSERT INTO foreign_constraints(
		foreign_stream,
		foreign_stream_id,
		stream,
		stream_id,
		saga_id
	) VALUES(
		$1, $2, $3, $4, $5
	) RETURNING *
	`

	DeleteForeignConstraintQuery = `
	DELETE FROM foreign_constraints
	WHERE foreign_stream = $1 AND foreign_stream_id = $2
	 AND stream = $3 AND stream_id = $4
	RETURNING *
	`

	ListAssociatedObjectsQuery = `
	SELECT stream, stream_id FROM foreign_constraints 
	WHERE foreign_stream = $1 AND foreign_stream = $2
	`
)
