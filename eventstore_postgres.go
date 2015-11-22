package eventhorizon

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// ErrCouldNotSaveEvent returned when an event could not be saved.
var ErrCouldNotSaveEvent = errors.New("could not save event")

// PostgresEventStore implements an EventStore for Postgres.
type PostgresEventStore struct {
	eventBus  EventBus
	db        *sqlx.DB
	factories map[string]func() Event
}

type postgresAggregateRecord struct {
	AggregateID string `db:"id"`
	Version     int
}

type postgresEventRecord struct {
	AggregrateID string
	Type         string
	Version      int
	Timestamp    time.Time
	Data         []byte
}

// NewPostgresEventStore creates a new PostgresEventStore.
func NewPostgresEventStore(eventBus EventBus, conn string) (*PostgresEventStore, error) {
	db, err := initDB(conn)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(`
CREATE TABLE IF NOT EXISTS aggregrates (
  id uuid NOT NULL,
  version int NOT NULL
);

CREATE TABLE IF NOT EXISTS events(
  aggregrateid uuid NOT NULL,
  type text,
  version int,
  timestamp timestamp without time zone default (now() at time zone 'utc'),
  data jsonb
)
    `); err != nil {
		db.Close()
		fmt.Println(err)
		return nil, ErrCouldNotCreateTables
	}

	return &PostgresEventStore{
		eventBus:  eventBus,
		db:        db,
		factories: make(map[string]func() Event),
	}, nil
}

// Save appends all events in the event stream to the store.
func (s *PostgresEventStore) Save(events []Event) error {
	if len(events) == 0 {
		return ErrNoEventsToAppend
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, event := range events {
		// Get an existing aggregate, if any
		var existing []postgresAggregateRecord
		err := s.db.Select(&existing,
			`SELECT * FROM aggregrates WHERE id=$1 LIMIT 2`, event.AggregateID().String())
		if (err != nil && err != sql.ErrNoRows) || len(existing) > 1 {
			return ErrCouldNotLoadAggregate
		}

		// Marshal event data
		b, err := json.Marshal(event)
		if err != nil {
			return ErrCouldNotMarshalEvent
		}

		// Create the event record with timestamp
		r := &postgresEventRecord{
			AggregrateID: event.AggregateID().String(),
			Type:         event.EventType(),
			Version:      1,
			Timestamp:    time.Now(),
			Data:         b,
		}

		if len(existing) == 0 {
			aggregrate := postgresAggregateRecord{
				AggregateID: event.AggregateID().String(),
				Version:     1,
			}

			// Save aggregrate
			_, err = s.db.NamedExec(
				`INSERT INTO aggregrates (id,version)
        VALUES (:id,:version)`, aggregrate)
			if err != nil {
				return ErrCouldNotSaveAggregate
			}

			_, err = s.db.NamedExec(
				`INSERT INTO events (aggregrateid,type,version,timestamp,data)
        VALUES (:aggregrateid,:type,:version,:timestamp,:data)`, r)
			if err != nil {
				return ErrCouldNotSaveEvent
			}
		} else {
			// Increment record version before inserting.
			version := existing[0].Version + 1
			_, err = s.db.NamedExec(
				`UPDATE aggregrates SET version=:version WHERE id=:id`, existing[0])
			if err != nil {
				return ErrCouldNotSaveAggregate
			}
			r.Version = version
			_, err = s.db.NamedExec(
				`INSERT INTO events (aggregrateid,type,version,timestamp,data)
        VALUES (:aggregrateid,:type,:version,:timestamp,:data)`, r)
			if err != nil {
				return ErrCouldNotSaveEvent
			}
		}

	}

	if err := tx.Commit(); err != nil {
		return err
	}

	for _, event := range events {
		// Publish event on the bus.
		if s.eventBus != nil {
			s.eventBus.PublishEvent(event)
		}
	}

	return nil
}

// Load loads all events for the aggregate id from the store.
func (s *PostgresEventStore) Load(id UUID) ([]Event, error) {
	var aggregrate postgresAggregateRecord
	err := s.db.Get(&aggregrate,
		`SELECT * FROM aggregrates WHERE id=$1 LIMIT 1`, id.String())
	if err != nil {
		return nil, ErrNoEventsFound
	}

	var rawEvents []*postgresEventRecord
	err = s.db.Select(&rawEvents,
		`SELECT * FROM events WHERE aggregrateid=$1 ORDER BY timestamp ASC`, id.String())
	if err != nil {
		return nil, ErrNoEventsFound
	}

	events := make([]Event, len(rawEvents))
	for i, rawEvent := range rawEvents {
		// Get the registered factory function for creating events.
		f, ok := s.factories[rawEvent.Type]
		if !ok {
			return nil, ErrEventNotRegistered
		}

		// Unmarshal JSON
		event := f()
		if err := json.Unmarshal(rawEvent.Data, event); err != nil {
			return nil, ErrCouldNotUnmarshalEvent
		}
		if events[i], ok = event.(Event); !ok {
			return nil, ErrInvalidEvent
		}

		rawEvent.Data = nil
	}

	if len(events) == 0 {
		events = nil
	}

	return events, nil
}

// RegisterEventType registers an event factory for a event type. The factory is
// used to create concrete event types when loading from the database.
//
// An example would be:
//     eventStore.RegisterEventType(&MyEvent{}, func() Event { return &MyEvent{} })
func (s *PostgresEventStore) RegisterEventType(event Event, factory func() Event) error {
	if _, ok := s.factories[event.EventType()]; ok {
		return ErrHandlerAlreadySet
	}

	s.factories[event.EventType()] = factory

	return nil
}

// Clear clears the postgres storage.
func (s *PostgresEventStore) Clear() error {
	if _, err := s.db.Exec(`DELETE FROM events`); err != nil {
		return ErrCouldNotClearDB
	}
	if _, err := s.db.Exec(`DELETE FROM aggregrates`); err != nil {
		return ErrCouldNotClearDB
	}

	return nil
}

// Close closes the postgres db connection.
func (s *PostgresEventStore) Close() error {
	return s.db.Close()
}
