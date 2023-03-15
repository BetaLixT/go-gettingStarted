// Package entities containing all data access objects (models that relate to
// the how the data is stored in the database)
package entities

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/events"
	"techunicorn.com/udc-core/gettingStarted/pkg/domain/base/foreigns"
	"techunicorn.com/udc-core/gettingStarted/pkg/infra/psqldb"

	// blank import to load postgresql drivers
	_ "github.com/lib/pq"
)

// =============================================================================
// Base Event DAOs
// =============================================================================

// BaseEvent representing the event model that every event should embed
type BaseEvent struct {
	ID        uint64    `db:"id"`
	SagaID    *string   `db:"saga_id"`
	Stream    string    `db:"stream"`
	StreamID  string    `db:"stream_id"`
	Event     string    `db:"event"`
	Version   uint64    `db:"version"`
	EventTime time.Time `db:"event_time"`
	TraceID   string    `db:"trace_id"`
	RequestID string    `db:"request_id"`
}

// GetID getter for ID
func (dao *BaseEvent) GetID() uint64 {
	return dao.ID
}

// GetEventTime getter for event time
func (dao *BaseEvent) GetEventTime() time.Time {
	return dao.EventTime
}

// ToDTO getting dto from dao structure
func (dao *BaseEvent) ToDTO() *events.EventEntity {
	return &events.EventEntity{
		Id:        dao.ID,
		SagaId:    dao.SagaID,
		Stream:    dao.Stream,
		StreamId:  dao.StreamID,
		Event:     dao.Event,
		Version:   dao.Version,
		EventTime: dao.EventTime,
	}
}

// IBaseEvent interface base base entity
type IBaseEvent interface {
	GetID() uint64
	// GetSagaID() *string
	// GetStream() string
	// GetStreamID() string
	// GetEvent() string
	// GetVersion() uint64
	GetEventTime() time.Time
	// GetTraceID() string
	// GetRequestID() string
}

// =============================================================================
// Uniques DAOs
// =============================================================================

// Unique dao for unique constraint
type Unique struct {
	Stream   string  `db:"stream"`
	StreamID string  `db:"stream_id"`
	SagaID   *string `db:"saga_id"`
	Property string  `db:"property"`
	Value    string  `db:"value"`
}

// =============================================================================
// ACL DAOs
// =============================================================================

// ACL dao representing an ACL entry
type ACL struct {
	Stream      string `db:"stream"`
	StreamID    string `db:"stream_id"`
	UserType    string `db:"user_type"`
	UserId      string `db:"user_id"`
	Permissions int    `db:"permissions"`
}

// =============================================================================
// Foreigns DAOs
// =============================================================================

// Foreign dao that represents a foreign entity
type Foreign struct {
	Stream   string  `db:"stream"`
	StreamID string  `db:"stream_id"`
	SagaID   *string `db:"saga_id"`
}

// ForeignConstraint dao that represents a foreign constraint entry
type ForeignConstraint struct {
	ForeignStream   string  `db:"foreign_stream"`
	ForeignStreamID string  `db:"foreign_stream_id"`
	Stream          string  `db:"stream"`
	StreamID        string  `db:"stream_id"`
	SagaID          *string `db:"saga_id"`
}

// ForeignAssociatedObject dao that represents a foreign constraint entry
type ForeignAssociatedObject struct {
	Stream   string `db:"stream"`
	StreamID string `db:"stream_id"`
}

func (dao *ForeignAssociatedObject) ToDTO() *foreigns.Object {
	return &foreigns.Object{
		Stream:   dao.Stream,
		StreamId: dao.StreamID,
	}
}

func (_ *ForeignAssociatedObject) ToDTOSlice(
	daos []ForeignAssociatedObject,
) []foreigns.Object {
	dtos := make([]foreigns.Object, len(daos))
	for idx := range daos {
		dtos[idx] = *daos[idx].ToDTO()
	}
	return dtos
}

// =============================================================================
// Common DAO models
// =============================================================================

// JSONObj dao for objects to be stored as json
type JSONObj map[string]interface{}

var _ driver.Value = (*JSONObj)(nil)

// Value for db writes
func (a JSONObj) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan for db reads
func (a *JSONObj) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// JSONMapString  dao for map[string]string to be stored as json
type JSONMapString map[string]string

var _ driver.Value = (*JSONObj)(nil)

// Value fo db writes
func (a JSONMapString) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan for db reads
func (a *JSONMapString) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// =============================================================================
// Migrations
// =============================================================================

// GetMigrationScripts provides all the migration scripts required for the
// application
func GetMigrationScripts() []psqldb.MigrationScript {
	migrationScripts := []psqldb.MigrationScript{
		{
			Key: "initial-event-source",
			Up: `
				CREATE TABLE events (
					id bigserial PRIMARY KEY NOT NULL,
					saga_id text,
					stream text NOT NULL,
					stream_id text NOT NULL,
					version bigint NOT NULL,
					event text NOT NULL,
					event_time timestamp with time zone NOT NULL,
					data bytea NOT NULL,
					trace_id text NOT NULL,
					request_id text NOT NULL,
					CONSTRAINT source_unique UNIQUE (stream, stream_id, version)
				);

				CREATE INDEX idx_events_stream_events ON events(stream, stream_id);

				CREATE TRIGGER set_events_event_time
				BEFORE INSERT ON events
				FOR EACH ROW
				EXECUTE PROCEDURE trigger_set_event_time();

				CREATE TABLE uniques (
					stream text NOT NULL,
					stream_id text NOT NULL,
					saga_id text,
					property text NOT NULL,
					value text NOT NULL,
					PRIMARY KEY(stream, property, value)
				);
				
				CREATE TABLE acl (
					stream text NOT NULL,
					stream_id text NOT NULL,
					user_type text NOT NULL,
					user_id text NOT NULL,
					permissions int NOT NULL,	
					PRIMARY KEY(stream, stream_id, user_type, user_id)
				);

				CREATE TABLE foreigns (
					stream text NOT NULL,
					stream_id text NOT NULL,
					saga_id text,
					PRIMARY KEY(stream, stream_id)
				);

				CREATE TABLE foreign_constraints (
					foreign_stream text NOT NULL,
					foreign_stream_id text NOT NULL,
					stream text NOT NULL,
					stream_id text NOT NULL,
					saga_id text,
					PRIMARY KEY (foreign_stream, foreign_stream_id, stream, stream_id),
					CONSTRAINT foreign_constraints_fk FOREIGN KEY (foreign_stream, foreign_stream_id) REFERENCES foreigns (stream, stream_id)
				);
				`,
			Down: `
			  DROP TABLE foreign_constraints;
			  DROP TABLE foreigns;
			  DROP TABLE acl;
			  DROP TABLE uniques;
				DROP TRIGGER set_events_event_time on events;
				DROP INDEX idx_events_stream_events;
				DROP TABLE events;
				`,
		},
	}
	return migrationScripts
}
