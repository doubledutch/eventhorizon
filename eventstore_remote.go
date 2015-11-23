package eventhorizon

import "errors"

// ErrCouldNotDialDB returned when the database could not be dialed.
var ErrCouldNotDialDB = errors.New("could not dial database")

// ErrNoDBSession returned when no database session is set.
var ErrNoDBSession = errors.New("no database session")

// ErrCouldNotClearDB returned when the database could not be cleared.
var ErrCouldNotClearDB = errors.New("could not clear database")

// ErrCouldNotMarshalEvent returned when an event could not be marshaled into BSON.
var ErrCouldNotMarshalEvent = errors.New("could not marshal event")

// ErrCouldNotUnmarshalEvent returned when an event could not be unmarshaled into a concrete type.
var ErrCouldNotUnmarshalEvent = errors.New("could not unmarshal event")

// RemoteEventStore is a store that is remote, requiring serialization and closing
type RemoteEventStore interface {
	EventStore
	RemoteHandler
	// Close the event store.
	Close() error
	// Clear deletes all data from the event store.
	Clear() error
}
