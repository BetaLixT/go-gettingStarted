// Package cassdb contains constructors, options and tracing plugins for gocql,
// a Cassandra Database client library
package cassdb

import (
	"context"
	"strconv"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/logger"

	trace "github.com/BetaLixT/appInsightsTrace"
	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

// NewCassandraSession constructs a new cassandra db client session
func NewCassandraSession(
	optn *Options,
	lgrf logger.IFactory,
	isn *trace.AppInsightsCore,
) (*gocql.Session, error) {
	lgr := lgrf.Create(context.Background())
	lgr.Info("connecting to cassandra...", zap.Strings("urls", optn.ClusterIPs))
	auth := gocql.PasswordAuthenticator{
		Username: optn.Username,
		Password: optn.Password,
	}

	// - creating main session
	cls := gocql.NewCluster(optn.ClusterIPs...)
	cls.Authenticator = auth
	cls.Keyspace = optn.Keyspace
	cls.QueryObserver = &traceObserver{
		ins: isn,
	}
	sess, err := cls.CreateSession()
	if err != nil {
		return nil, err
	}

	return sess, nil
}

type traceObserver struct {
	ins *trace.AppInsightsCore
}

func (o *traceObserver) ObserveQuery(ctx context.Context, qry gocql.ObservedQuery) {
	o.ins.TraceDependency(
		ctx,
		"",
		"cassandra",
		qry.Host.ClusterName(),
		qry.Statement,
		qry.Err == nil,
		qry.Start,
		qry.End,
		map[string]string{
			"keyspace":    qry.Keyspace,
			"serverLtncy": strconv.FormatInt(qry.Metrics.TotalLatency, 10),
		},
	)
}
