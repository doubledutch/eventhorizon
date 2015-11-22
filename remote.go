package eventhorizon

import "errors"

// ErrEventNotRegistered returned when an event is not registered.
var ErrEventNotRegistered = errors.New("event not registered")

// RemoteHandler enables deserizliaing remote objects to events
type RemoteHandler interface {
	// Register a function to create a new event from an event
	RegisterEventType(Event, func() Event) error
}
