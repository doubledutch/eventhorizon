package eventhorizon

import "time"

// MemoryEventStore implements EventStore as an in memory structure.
type MemoryEventStore struct {
	eventBus         EventBus
	aggregateRecords map[UUID]*memoryAggregateRecord
}

// NewMemoryEventStore creates a new MemoryEventStore.
func NewMemoryEventStore(eventBus EventBus) *MemoryEventStore {
	s := &MemoryEventStore{
		eventBus:         eventBus,
		aggregateRecords: make(map[UUID]*memoryAggregateRecord),
	}
	return s
}

// Save appends all events in the event stream to the memory store.
func (s *MemoryEventStore) Save(events []Event) error {
	if len(events) == 0 {
		return ErrNoEventsToAppend
	}

	for _, event := range events {
		r := &memoryEventRecord{
			eventType: event.EventType(),
			timestamp: time.Now(),
			event:     event,
		}

		if a, ok := s.aggregateRecords[event.AggregateID()]; ok {
			a.version++
			r.version = a.version
			a.events = append(a.events, r)
		} else {
			s.aggregateRecords[event.AggregateID()] = &memoryAggregateRecord{
				aggregateID: event.AggregateID(),
				version:     0,
				events:      []*memoryEventRecord{r},
			}
		}

		// Publish event on the bus.
		if s.eventBus != nil {
			s.eventBus.PublishEvent(event)
		}
	}

	return nil
}

// Load loads all events for the aggregate id from the memory store.
// Returns ErrNoEventsFound if no events can be found.
func (s *MemoryEventStore) Load(id UUID) ([]Event, error) {
	if a, ok := s.aggregateRecords[id]; ok {
		events := make([]Event, len(a.events))
		for i, r := range a.events {
			events[i] = r.event
		}
		return events, nil
	}

	return nil, ErrNoEventsFound
}

// Close closes the store.
func (s *MemoryEventStore) Close() error {
	return nil
}

type memoryAggregateRecord struct {
	aggregateID UUID
	version     int
	events      []*memoryEventRecord
}

type memoryEventRecord struct {
	eventType string
	version   int
	timestamp time.Time
	event     Event
}
