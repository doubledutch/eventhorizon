package eventhorizon

import (
	. "gopkg.in/check.v1"
)

var _ = Suite(&MemoryEventStoreSuite{})

type MemoryEventStoreSuite struct {
	EventStoreSuite
}

func (s *MemoryEventStoreSuite) SetUpTest(c *C) {
	bus := &MockEventBus{
		events: make([]Event, 0),
	}
	s.Store = NewMemoryEventStore(bus)
}

func (s *MemoryEventStoreSuite) Test_NewMemoryEventStore(c *C) {
	bus := &MockEventBus{
		events: make([]Event, 0),
	}
	store := NewMemoryEventStore(bus)
	c.Assert(store, NotNil)
}
