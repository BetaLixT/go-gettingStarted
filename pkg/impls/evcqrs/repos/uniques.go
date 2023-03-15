package repos

import (
	"context"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/uniques"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/common"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/entities"

	"go.uber.org/zap"
)

type UniquesRepository struct {
	*BaseDataRepository
	lgrf logger.IFactory
}

var _ uniques.IRepository = (*UniquesRepository)(nil)

func NewUniquesRepository(
	base *BaseDataRepository,
	lgrf logger.IFactory,
) *UniquesRepository {
	return &UniquesRepository{
		BaseDataRepository: base,
		lgrf:               lgrf,
	}
}

func (r *UniquesRepository) RegisterConstraint(
	c context.Context,
	stream string,
	streamId string,
	sagaId *string,
	property string,
	value string,
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

	unqDao := entities.Unique{}
	err = dbtx.Get(
		ctx,
		&unqDao,
		InsertConstraintQuery,
		stream,
		streamId,
		sagaId,
		property,
		value,
	)
	if err != nil {
		lgr.Error("failed to insert unique constraint",
			zap.Error(err),
		)
	}

	return err
}

func (r *UniquesRepository) RemoveConstraint(
	c context.Context,
	stream string,
	property string,
	value string,
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

	// deleting constraint
	unqDao := entities.Unique{}
	err = dbtx.Get(
		ctx,
		&unqDao,
		DeleteConstraintQuery,
		stream,
		property,
		value,
	)
	if err != nil {
		lgr.Error("failed to delete unique constraint",
			zap.Error(err),
		)
	}

	return err
}

const (
	InsertConstraintQuery = `
	INSERT INTO uniques(
		stream,
		stream_id,
		saga_id,
		property,
		value
	) VALUES(
		$1, $2, $3, $4, $5
	) RETURNING *
	`
	DeleteConstraintQuery = `
	DELETE FROM uniques
	WHERE stream = $1 AND property = $2 AND value = $3
	RETURNING *
	`
)
