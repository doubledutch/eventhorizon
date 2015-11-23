package eventhorizon

import (
	"github.com/odeke-em/go-uuid"

	. "gopkg.in/check.v1"
)

type TestCommandHandler struct {
	command Command
	recv    chan struct{}
}

func (t *TestCommandHandler) HandleCommand(command Command) error {
	t.command = command
	if t.recv != nil {
		t.recv <- struct{}{}
	}
	return nil
}

type CommandBusSuite struct {
	bus CommandBus
}

func (s *CommandBusSuite) Setup(bus CommandBus) {
	s.bus = bus
}

func (s *CommandBusSuite) Test_HandlePublish_Simple(c *C) {
	handler := &TestCommandHandler{
		recv: make(chan struct{}, 1),
	}
	err := s.bus.SetHandler(handler, &TestCommand{})
	c.Assert(err, IsNil)
	command1 := &TestCommand{uuid.New(), "command1"}
	err = s.bus.PublishCommand(command1)
	c.Assert(err, IsNil)
	<-handler.recv
	c.Assert(handler.command, DeepEquals, command1)
}

func (s *CommandBusSuite) Test_HandleCommand_Simple(c *C) {
	handler := &TestCommandHandler{}
	err := s.bus.SetHandler(handler, &TestCommand{})
	c.Assert(err, IsNil)
	command1 := &TestCommand{uuid.New(), "command1"}
	err = s.bus.HandleCommand(command1)
	c.Assert(err, IsNil)
	c.Assert(handler.command, Equals, command1)
}

func (s *CommandBusSuite) Test_HandleCommand_NoHandler(c *C) {
	handler := &TestCommandHandler{}
	command1 := &TestCommand{uuid.New(), "command1"}
	err := s.bus.HandleCommand(command1)
	c.Assert(err, Equals, ErrHandlerNotFound)
	c.Assert(handler.command, IsNil)
}

func (s *CommandBusSuite) Test_SetHandler_Twice(c *C) {
	handler := &TestCommandHandler{}
	err := s.bus.SetHandler(handler, &TestCommand{})
	c.Assert(err, IsNil)
	err = s.bus.SetHandler(handler, &TestCommand{})
	c.Assert(err, Equals, ErrHandlerAlreadySet)
}
