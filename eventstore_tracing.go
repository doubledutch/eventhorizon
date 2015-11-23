package eventhorizon

// TraceEventStore wraps an EventStore and adds debug tracing.
type TraceEventStore struct {
	eventStore EventStore
	tracing    bool
	trace      []Event
}

// NewTraceEventStore creates a new TraceEventStore.
func NewTraceEventStore(eventStore EventStore) *TraceEventStore {
	s := &TraceEventStore{
		eventStore: eventStore,
		trace:      make([]Event, 0),
	}
	return s
}

// Save appends all events to the base store and trace them if enabled.
func (s *TraceEventStore) Save(events []Event) error {
	if s.tracing {
		s.trace = append(s.trace, events...)
	}

	if s.eventStore != nil {
		return s.eventStore.Save(events)
	}

	return nil
}

// Load loads all events for the aggregate id from the base store.
// Returns ErrNoEventStoreDefined if no event store could be found.
func (s *TraceEventStore) Load(id UUID) ([]Event, error) {
	if s.eventStore != nil {
		return s.eventStore.Load(id)
	}

	return nil, ErrNoEventStoreDefined
}

// StartTracing starts the tracing of events.
func (s *TraceEventStore) StartTracing() {
	s.tracing = true
}

// StopTracing stops the tracing of events.
func (s *TraceEventStore) StopTracing() {
	s.tracing = false
}

// GetTrace returns the events that happened during the tracing.
func (s *TraceEventStore) GetTrace() []Event {
	return s.trace
}

// ResetTrace resets the trace.
func (s *TraceEventStore) ResetTrace() {
	s.trace = make([]Event, 0)
}
