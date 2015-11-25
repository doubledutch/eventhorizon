// Copyright (c) 2014 - Max Persson <max@looplab.se>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eventhorizon

import (
	"fmt"
	"time"

	"github.com/odeke-em/go-uuid"

	. "gopkg.in/check.v1"
)

var _ = Suite(&AggregateCommandHandlerSuite{})

type AggregateCommandHandlerSuite struct {
	repo    *MockRepository
	handler *AggregateCommandHandler
}

func (s *AggregateCommandHandlerSuite) SetUpTest(c *C) {
	s.repo = &MockRepository{
		aggregates: make(map[string]Aggregate),
	}
	s.handler, _ = NewAggregateCommandHandler(s.repo)
}

func (s *AggregateCommandHandlerSuite) Test_NewDispatcher(c *C) {
	repo := &MockRepository{
		aggregates: make(map[string]Aggregate),
	}
	handler, err := NewAggregateCommandHandler(repo)
	c.Assert(handler, NotNil)
	c.Assert(err, IsNil)
}

func (s *AggregateCommandHandlerSuite) Test_NewDispatcher_ErrNilRepository(c *C) {
	handler, err := NewAggregateCommandHandler(nil)
	c.Assert(handler, IsNil)
	c.Assert(err, Equals, ErrNilRepository)
}

var dispatchedCommand Command

type TestDispatcherAggregate struct {
	*AggregateBase
}

func (t *TestDispatcherAggregate) AggregateType() string {
	return "TestDispatcherAggregate"
}

func (t *TestDispatcherAggregate) HandleCommand(command Command) error {
	dispatchedCommand = command
	switch command := command.(type) {
	case *TestCommand:
		if command.Content == "error" {
			return fmt.Errorf("command error")
		}
		t.StoreEvent(&TestEvent{command.TestID, command.Content})
		return nil
	}
	return fmt.Errorf("couldn't handle command")
}

func (t *TestDispatcherAggregate) ApplyEvent(event Event) {
}

func (s *AggregateCommandHandlerSuite) Test_Simple(c *C) {
	aggregate := &TestDispatcherAggregate{
		AggregateBase: NewAggregateBase(uuid.New()),
	}
	s.repo.aggregates[aggregate.AggregateID()] = aggregate
	s.handler.SetAggregate(aggregate, &TestCommand{})
	command1 := &TestCommand{aggregate.AggregateID(), "command1"}
	err := s.handler.HandleCommand(command1)
	c.Assert(dispatchedCommand, Equals, command1)
	c.Assert(err, IsNil)
}

func (s *AggregateCommandHandlerSuite) Test_ErrorInHandler(c *C) {
	aggregate := &TestDispatcherAggregate{
		AggregateBase: NewAggregateBase(uuid.New()),
	}
	s.repo.aggregates[aggregate.AggregateID()] = aggregate
	s.handler.SetAggregate(aggregate, &TestCommand{})
	commandError := &TestCommand{aggregate.AggregateID(), "error"}
	err := s.handler.HandleCommand(commandError)
	c.Assert(err, ErrorMatches, "command error")
	c.Assert(dispatchedCommand, Equals, commandError)
}

func (s *AggregateCommandHandlerSuite) Test_NoHandlers(c *C) {
	command1 := &TestCommand{uuid.New(), "command1"}
	err := s.handler.HandleCommand(command1)
	c.Assert(err, Equals, ErrAggregateNotFound)
}

func (s *AggregateCommandHandlerSuite) Test_SetHandler_Twice(c *C) {
	aggregate := &TestDispatcherAggregate{}
	err := s.handler.SetAggregate(aggregate, &TestCommand{})
	c.Assert(err, IsNil)
	aggregate2 := &TestDispatcherAggregate{}
	err = s.handler.SetAggregate(aggregate2, &TestCommand{})
	c.Assert(err, Equals, ErrAggregateAlreadySet)
}

var callCountDispatcher int

type BenchmarkDispatcherAggregate struct {
	*AggregateBase
}

func (t *BenchmarkDispatcherAggregate) AggregateType() string {
	return "BenchmarkDispatcherAggregate"
}

func (t *BenchmarkDispatcherAggregate) HandleCommand(command Command) error {
	callCountDispatcher++
	return nil
}

func (t *BenchmarkDispatcherAggregate) ApplyEvent(event Event) {
}

func (s *AggregateCommandHandlerSuite) Benchmark_Dispatcher(c *C) {
	repo := &MockRepository{
		aggregates: make(map[string]Aggregate),
	}
	handler, _ := NewAggregateCommandHandler(repo)
	agg := &TestDispatcherAggregate{
		AggregateBase: NewAggregateBase(uuid.New()),
	}
	repo.aggregates[agg.AggregateID()] = agg
	handler.SetAggregate(agg, &TestCommand{})

	callCountDispatcher = 0
	command1 := &TestCommand{agg.AggregateID(), "command1"}
	for i := 0; i < c.N; i++ {
		handler.HandleCommand(command1)
	}
	c.Assert(callCountDispatcher, Equals, c.N)
}

func (s *AggregateCommandHandlerSuite) Test_CheckCommand_AllFields(c *C) {
	err := s.handler.checkCommand(&TestCommand{uuid.New(), "command1"})
	c.Assert(err, Equals, nil)
}

type TestCommandValue struct {
	TestID  string
	Content string
}

func (t *TestCommandValue) AggregateID() string   { return t.TestID }
func (t *TestCommandValue) AggregateType() string { return "Test" }
func (t *TestCommandValue) CommandType() string   { return "TestCommandValue" }

func (s *AggregateCommandHandlerSuite) Test_CheckCommand_MissingRequired_Value(c *C) {
	err := s.handler.checkCommand(&TestCommandValue{TestID: uuid.New()})
	c.Assert(err, ErrorMatches, "missing field: Content")
}

type TestCommandSlice struct {
	TestID string
	Slice  []string
}

func (t *TestCommandSlice) AggregateID() string   { return t.TestID }
func (t *TestCommandSlice) AggregateType() string { return "Test" }
func (t *TestCommandSlice) CommandType() string   { return "TestCommandSlice" }

func (s *AggregateCommandHandlerSuite) Test_CheckCommand_MissingRequired_Slice(c *C) {
	err := s.handler.checkCommand(&TestCommandSlice{TestID: uuid.New()})
	c.Assert(err, ErrorMatches, "missing field: Slice")
}

type TestCommandMap struct {
	TestID string
	Map    map[string]string
}

func (t *TestCommandMap) AggregateID() string   { return t.TestID }
func (t *TestCommandMap) AggregateType() string { return "Test" }
func (t *TestCommandMap) CommandType() string   { return "TestCommandMap" }

func (s *AggregateCommandHandlerSuite) Test_CheckCommand_MissingRequired_Map(c *C) {
	err := s.handler.checkCommand(&TestCommandMap{TestID: uuid.New()})
	c.Assert(err, ErrorMatches, "missing field: Map")
}

type TestCommandStruct struct {
	TestID string
	Struct struct {
		Test string
	}
}

func (t *TestCommandStruct) AggregateID() string   { return t.TestID }
func (t *TestCommandStruct) AggregateType() string { return "Test" }
func (t *TestCommandStruct) CommandType() string   { return "TestCommandStruct" }

func (s *AggregateCommandHandlerSuite) Test_CheckCommand_MissingRequired_Struct(c *C) {
	err := s.handler.checkCommand(&TestCommandStruct{TestID: uuid.New()})
	c.Assert(err, ErrorMatches, "missing field: Struct")
}

type TestCommandTime struct {
	TestID string
	Time   time.Time
}

func (t *TestCommandTime) AggregateID() string   { return t.TestID }
func (t *TestCommandTime) AggregateType() string { return "Test" }
func (t *TestCommandTime) CommandType() string   { return "TestCommandTime" }

func (s *AggregateCommandHandlerSuite) Test_CheckCommand_MissingRequired_Time(c *C) {
	err := s.handler.checkCommand(&TestCommandTime{TestID: uuid.New()})
	c.Assert(err, ErrorMatches, "missing field: Time")
}

type TestCommandOptional struct {
	TestID  string
	Content string `eh:"optional"`
}

func (t *TestCommandOptional) AggregateID() string   { return t.TestID }
func (t *TestCommandOptional) AggregateType() string { return "Test" }
func (t *TestCommandOptional) CommandType() string   { return "TestCommandOptional" }

func (s *AggregateCommandHandlerSuite) Test_CheckCommand_MissingOptionalField(c *C) {
	err := s.handler.checkCommand(&TestCommandOptional{TestID: uuid.New()})
	c.Assert(err, Equals, nil)
}

type TestCommandPrivate struct {
	TestID  string
	private string
}

func (t *TestCommandPrivate) AggregateID() string   { return t.TestID }
func (t *TestCommandPrivate) AggregateType() string { return "Test" }
func (t *TestCommandPrivate) CommandType() string   { return "TestCommandPrivate" }

func (s *AggregateCommandHandlerSuite) Test_CheckCommand_MissingPrivateField(c *C) {
	err := s.handler.checkCommand(&TestCommandPrivate{TestID: uuid.New()})
	c.Assert(err, Equals, nil)
}
