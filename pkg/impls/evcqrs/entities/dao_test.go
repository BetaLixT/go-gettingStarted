package entities

import (
	"context"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/psqldb"
	"testing"
	"time"

	"go.uber.org/zap"
)

type MockTracer struct {
}

func (t *MockTracer) TraceDependency(
	ctx context.Context,
	spanId string,
	dependencyType string,
	serviceName string,
	commandName string,
	success bool,
	startTimestamp time.Time,
	eventTimestamp time.Time,
	fields map[string]string,
) {

}

const (
	TestDatabaseConnString  = "host=127.0.0.1 port=5432 user=admin password=123456 dbname=todo_test sslmode=disable"
	TestDatabaseServiceName = "test-db"
)

func TestMigration(t *testing.T) {
	dbctx, err := psqldb.NewDatabaseContext(
		&MockTracer{},
		&psqldb.DatabaseOptions{
			ConnectionString:    TestDatabaseConnString,
			DatabaseServiceName: TestDatabaseServiceName,
		},
	)
	if err != nil {
		println("failed to connect to database, skipping test")
		t.SkipNow()
	}
	ctx := context.Background()
	lgr, _ := zap.NewDevelopment()
	err = psqldb.RunMigrations(
		ctx,
		lgr,
		dbctx,
		GetMigrationScripts(),
	)
	if err != nil {
		println("failed to run migration")
		t.FailNow()
	}
}
