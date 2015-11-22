package eventhorizon

import (
	. "gopkg.in/check.v1"
)

type EventStoreSuite struct {
	Store EventStore
	Bus   EventBus
}

func (s *EventStoreSuite) Test_NoEvents(c *C) {
	err := s.Store.Save([]Event{})
	c.Assert(err, Equals, ErrNoEventsToAppend)
}

func (s *EventStoreSuite) Test_OneEvent(c *C) {
	event1 := &TestEvent{NewUUID(), "event1"}
	err := s.Store.Save([]Event{event1})
	c.Assert(err, IsNil)
	events, err := s.Store.Load(event1.TestID)
	c.Assert(err, IsNil)
	c.Assert(events, HasLen, 1)
	c.Assert(events[0], DeepEquals, event1)
}

func (s *EventStoreSuite) Test_TwoEvents(c *C) {
	event1 := &TestEvent{NewUUID(), "event1"}
	event2 := &TestEvent{event1.TestID, "event2"}
	err := s.Store.Save([]Event{event1, event2})
	c.Assert(err, IsNil)
	events, err := s.Store.Load(event1.TestID)
	c.Assert(err, IsNil)
	c.Assert(events, HasLen, 2)
	c.Assert(events[0], DeepEquals, event1)
	c.Assert(events[1], DeepEquals, event2)
}

func (s *EventStoreSuite) Test_DifferentAggregates(c *C) {
	event1 := &TestEvent{NewUUID(), "event1"}
	event2 := &TestEvent{NewUUID(), "event2"}
	err := s.Store.Save([]Event{event1, event2})
	c.Assert(err, IsNil)
	events, err := s.Store.Load(event1.TestID)
	c.Assert(err, IsNil)
	c.Assert(events, HasLen, 1)
	c.Assert(events[0], DeepEquals, event1)
	events, err = s.Store.Load(event2.TestID)
	c.Assert(err, IsNil)
	c.Assert(events, HasLen, 1)
	c.Assert(events[0], DeepEquals, event2)
}

func (s *EventStoreSuite) Test_LoadNoEvents(c *C) {
	events, err := s.Store.Load(NewUUID())
	c.Assert(err, ErrorMatches, "could not find events")
	c.Assert(events, DeepEquals, []Event(nil))
}
