package eventhorizon

import (
	. "gopkg.in/check.v1"
)

type EventBusSuite struct {
	Bus  EventBus
	Bus2 EventBus
}

func (s *EventBusSuite) Test_PublishEvent_Simple(c *C) {
	handler := NewMockEventHandler()
	localHandler := NewMockEventHandler()
	globalHandler := NewMockEventHandler()
	globalHandler2 := NewMockEventHandler()
	s.Bus.AddHandler(handler, &TestEvent{})
	s.Bus.AddLocalHandler(localHandler)
	s.Bus.AddGlobalHandler(globalHandler)
	s.Bus2.AddGlobalHandler(globalHandler2)

	event1 := &TestEvent{NewUUID(), "event1"}
	s.Bus.PublishEvent(event1)
	<-globalHandler.recv
	<-globalHandler2.recv
	c.Assert(handler.events, HasLen, 1)
	c.Assert(handler.events[0], DeepEquals, event1)
	c.Assert(localHandler.events, HasLen, 1)
	c.Assert(localHandler.events[0], DeepEquals, event1)
	c.Assert(globalHandler.events, HasLen, 1)
	c.Assert(globalHandler.events[0], DeepEquals, event1)
	c.Assert(globalHandler2.events, HasLen, 1)
	c.Assert(globalHandler2.events[0], DeepEquals, event1)
}

func (s *EventBusSuite) Test_PublishEvent_AnotherEvent(c *C) {
	handler := NewMockEventHandler()
	localHandler := NewMockEventHandler()
	globalHandler := NewMockEventHandler()
	globalHandler2 := NewMockEventHandler()
	s.Bus.AddHandler(handler, &TestEventOther{})
	s.Bus.AddLocalHandler(localHandler)
	s.Bus.AddGlobalHandler(globalHandler)
	s.Bus2.AddGlobalHandler(globalHandler2)

	event1 := &TestEvent{NewUUID(), "event1"}
	s.Bus.PublishEvent(event1)
	<-globalHandler.recv
	<-globalHandler2.recv
	c.Assert(handler.events, HasLen, 0)
	c.Assert(localHandler.events, HasLen, 1)
	c.Assert(localHandler.events[0], DeepEquals, event1)
	c.Assert(globalHandler.events, HasLen, 1)
	c.Assert(globalHandler.events[0], DeepEquals, event1)
	c.Assert(globalHandler2.events, HasLen, 1)
	c.Assert(globalHandler2.events[0], DeepEquals, event1)
}

func (s *EventBusSuite) Test_PublishEvent_NoHandler(c *C) {
	localHandler := NewMockEventHandler()
	globalHandler := NewMockEventHandler()
	globalHandler2 := NewMockEventHandler()
	s.Bus.AddLocalHandler(localHandler)
	s.Bus.AddGlobalHandler(globalHandler)
	s.Bus2.AddGlobalHandler(globalHandler2)

	event1 := &TestEvent{NewUUID(), "event1"}
	s.Bus.PublishEvent(event1)
	<-globalHandler.recv
	<-globalHandler2.recv
	c.Assert(localHandler.events, HasLen, 1)
	c.Assert(localHandler.events[0], DeepEquals, event1)
	c.Assert(globalHandler.events, HasLen, 1)
	c.Assert(globalHandler.events[0], DeepEquals, event1)
	c.Assert(globalHandler2.events, HasLen, 1)
	c.Assert(globalHandler2.events[0], DeepEquals, event1)
}

func (s *EventBusSuite) Test_PublishEvent_NoLocalOrGlobalHandler(c *C) {
	handler := NewMockEventHandler()
	s.Bus.AddHandler(handler, &TestEvent{})

	event1 := &TestEvent{NewUUID(), "event1"}
	s.Bus.PublishEvent(event1)
	c.Assert(handler.events, HasLen, 1)
	c.Assert(handler.events[0], DeepEquals, event1)
}
