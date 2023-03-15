// Package evcqrs Event source CQRS implementation of the domain layer
package evcqrs

import (
	"context"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/acl"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/cntxt"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/foreigns"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/impl"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/uids"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/uniques"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/entities"
	"techunicorn.com/udc-core/gettingStarted/pkg/impls/evcqrs/repos"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/lgr"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/psqldb"

	"github.com/BetaLixT/tsqlx"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// DependencySet dependencies provided by the implementation
var DependencySet = wire.NewSet(
	NewImplementation,
	wire.Bind(
		new(impl.IImplementation),
		new(*Implementation),
	),

	// Repos
	repos.NewBaseDataRepository,
	repos.NewACLRepository,
	wire.Bind(
		new(acl.IRepository),
		new(*repos.ACLRepository),
	),
	repos.NewContextFactory,
	wire.Bind(
		new(cntxt.IFactory),
		new(*repos.ContextFactory),
	),
	repos.NewForeignsRepository,
	wire.Bind(
		new(foreigns.IRepository),
		new(*repos.ForeignsRepository),
	),
	repos.NewUniquesRepository,
	wire.Bind(
		new(uniques.IRepository),
		new(*repos.UniquesRepository),
	),
	repos.NewUIDRepository,
	wire.Bind(
		new(uids.IRepository),
		new(*repos.UIDRepository),
	),
)

// Implementation used for graceful starting and stopping of the implementation
// layer
type Implementation struct {
	dbctx *tsqlx.TracedDB
	lgrf  *lgr.LoggerFactory
	rdb   *redis.Client
}

// NewImplementation constructor for the evcqrs implementation
func NewImplementation(
	dbctx *tsqlx.TracedDB,
	rdb *redis.Client,
	lgrf *lgr.LoggerFactory,
) *Implementation {
	return &Implementation{
		dbctx: dbctx,
		lgrf:  lgrf,
		rdb:   rdb,
	}
}

// Start runs any routines that are required before the implemtation layer can
// be utilized
func (i *Implementation) Start(ctx context.Context) error {
	lgri := i.lgrf.Create(ctx)
	err := psqldb.RunMigrations(
		ctx,
		lgri,
		i.dbctx,
		entities.GetMigrationScripts(),
	)
	if err != nil {
		lgri.Error("failed to run migration", zap.Error(err))
		return err
	}
	return nil
}

// Stop runs any routines that are required for the implementation layer to
// gracefully shutdown
func (i *Implementation) Stop(ctx context.Context) error {
	i.lgrf.Close()
	return nil
}

// StatusCheck checks connections with dependencies and returns error if any
// fail
func (i *Implementation) StatusCheck(ctx context.Context) error {
	lgr := i.lgrf.Create(ctx)

	lgr.Info("pinging psql database...")
	err := i.dbctx.Ping()
	if err != nil {
		lgr.Error("failed pinging database", zap.Error(err))
		return err
	}
	lgr.Info("psql ok")

	lgr.Info("pinging redis...")
	err = i.rdb.Ping(ctx).Err()
	if err != nil {
		lgr.Error("failed pinging redis")
		return err
	}
	lgr.Info("rdb ok")

	return nil
}
