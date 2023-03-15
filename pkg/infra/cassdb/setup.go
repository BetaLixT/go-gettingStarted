package cassdb

import (
	"context"
	"fmt"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

type CassandraSetup struct {
	sess *gocql.Session
	lgrf logger.IFactory
}

func NewCassandraSetup(
	sess *gocql.Session,
	lgrf logger.IFactory,
) *CassandraSetup {
	return &CassandraSetup{
		sess: sess,
		lgrf: lgrf,
	}
}

func (cs *CassandraSetup) Initialize(
	ctx context.Context,
) error {
	lgr := cs.lgrf.Create(ctx)
	err := cs.sess.Query(Schema).Exec()
	if err != nil {
		lgr.Error("error while initializing cassandra", zap.Error(err))
		return fmt.Errorf("error while initializing cassandra: %w", err)
	}
	return nil
}

const (
	Schema = `
  CREATE TABLE IF NOT EXISTS livestream_user (
    session_id text,
    participant_id text,
    id uuid,
    PRIMARY KEY ((session_id, participant_id))
  );
  `
)
