package eventhorizon

import (
	"github.com/odeke-em/go-uuid"
	. "gopkg.in/check.v1"
)

type EventBusSuite struct {
	Bus  EventBus
	Bus2 EventBus
}

func (s *EventBusSuite) Test_PublishEvent_Simple(c *C) {
	handler := NewMockEventHandler()
	defer handler.Close()
	localHandler := NewMockEventHandler()
	defer localHandler.Close()
	globalHandler := NewMockEventHandler()
	defer globalHandler.Close()
	s.Bus.AddHandler(handler, &TestEvent{})
	s.Bus.AddLocalHandler(localHandler)
	s.Bus.AddGlobalHandler(globalHandler)

	event1 := &TestEvent{uuid.New(), "event1"}
	s.Bus.PublishEvent(event1)
	<-globalHandler.recv
	c.Assert(handler.events, HasLen, 1)
	c.Assert(handler.events[0], DeepEquals, event1)
	c.Assert(localHandler.events, HasLen, 1)
	c.Assert(localHandler.events[0], DeepEquals, event1)
	c.Assert(globalHandler.events, HasLen, 1)
	c.Assert(globalHandler.events[0], DeepEquals, event1)
}

func (s *EventBusSuite) Test_PublishEvent_AnotherEvent(c *C) {
	handler := NewMockEventHandler()
	defer handler.Close()
	localHandler := NewMockEventHandler()
	defer localHandler.Close()
	globalHandler := NewMockEventHandler()
	defer globalHandler.Close()
	s.Bus.AddHandler(handler, &TestEventOther{})
	s.Bus.AddLocalHandler(localHandler)
	s.Bus.AddGlobalHandler(globalHandler)

	event1 := &TestEvent{uuid.New(), "event1"}
	s.Bus.PublishEvent(event1)
	<-globalHandler.recv
	event2 := &TestEvent{uuid.New(), "event2"}
	s.Bus.PublishEvent(event2)
	<-globalHandler.recv
	event3 := &TestEvent{uuid.New(), "event3"}
	s.Bus2.PublishEvent(event3)
	<-globalHandler.recv
	c.Assert(handler.events, HasLen, 0)
	c.Assert(localHandler.events, HasLen, 2)
	c.Assert(localHandler.events, DeepEquals, []Event{event1, event2})
	c.Assert(globalHandler.events, HasLen, 3)
	c.Assert(globalHandler.events, DeepEquals, []Event{event1, event2, event3})
}

func (s *EventBusSuite) Test_PublishEvent_NoHandler(c *C) {
	localHandler := NewMockEventHandler()
	defer localHandler.Close()
	globalHandler := NewMockEventHandler()
	defer globalHandler.Close()
	s.Bus.AddLocalHandler(localHandler)
	s.Bus.AddGlobalHandler(globalHandler)

	event1 := &TestEvent{uuid.New(), "event1"}
	s.Bus.PublishEvent(event1)
	<-globalHandler.recv
	c.Assert(localHandler.events, HasLen, 1)
	c.Assert(localHandler.events[0], DeepEquals, event1)
	c.Assert(globalHandler.events, HasLen, 1)
	c.Assert(globalHandler.events[0], DeepEquals, event1)
}

func (s *EventBusSuite) Test_PublishEvent_NoLocalOrGlobalHandler(c *C) {
	handler := NewMockEventHandler()
	defer handler.Close()
	s.Bus.AddHandler(handler, &TestEvent{})

	event1 := &TestEvent{uuid.New(), "event1"}
	s.Bus.PublishEvent(event1)
	<-handler.recv
	c.Assert(handler.events, HasLen, 1)
	c.Assert(handler.events[0], DeepEquals, event1)
}
