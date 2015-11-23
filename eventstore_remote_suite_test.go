package eventhorizon

import (
	. "gopkg.in/check.v1"
)

type RemoteEventStoreSuite struct {
	Store RemoteEventStore
	EventStoreSuite
}

func (s *RemoteEventStoreSuite) Setup(store RemoteEventStore, c *C) {
	err := store.RegisterEventType(&TestEvent{}, func() Event { return &TestEvent{} })
	c.Assert(err, IsNil)
	store.Clear()

	s.Store = store
	s.EventStoreSuite.Store = store
}

func (s *RemoteEventStoreSuite) TearDownTest(c *C) {
	err := s.Store.Close()
	c.Assert(err, IsNil)
}

func (s *RemoteEventStoreSuite) Test_NotRegisteredEvent(c *C) {
	event1 := &TestEventOther{NewUUID(), "event1"}
	err := s.Store.Save([]Event{event1})
	c.Assert(err, IsNil)
	events, err := s.Store.Load(event1.TestID)
	c.Assert(events, IsNil)
	c.Assert(err, Equals, ErrEventNotRegistered)
}
