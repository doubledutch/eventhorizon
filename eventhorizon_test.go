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
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
//
// Run benchmarks with "go test -check.b"
func Test(t *testing.T) { TestingT(t) }

type EmptyAggregate struct {
}

type TestAggregate struct {
	*AggregateBase
	events []Event
}

func (t *TestAggregate) AggregateType() string {
	return "TestAggregate"
}

func (t *TestAggregate) ApplyEvent(event Event) {
	t.events = append(t.events, event)
}

type TestEvent struct {
	TestID  string
	Content string
}

func (t *TestEvent) AggregateID() string   { return t.TestID }
func (t *TestEvent) AggregateType() string { return "Test" }
func (t *TestEvent) EventType() string     { return "TestEvent" }

type TestEventOther struct {
	TestID  string
	Content string
}

func (t *TestEventOther) AggregateID() string   { return t.TestID }
func (t *TestEventOther) AggregateType() string { return "Test" }
func (t *TestEventOther) EventType() string     { return "TestEventOther" }

type TestCommand struct {
	TestID  string
	Content string
}

func (t *TestCommand) AggregateID() string   { return t.TestID }
func (t *TestCommand) AggregateType() string { return "Test" }
func (t *TestCommand) CommandType() string   { return "TestCommand" }

type TestCommandOther struct {
	TestID  string
	Content string
}

func (t *TestCommandOther) AggregateID() string   { return t.TestID }
func (t *TestCommandOther) AggregateType() string { return "Test" }
func (t *TestCommandOther) CommandType() string   { return "TestCommandOther" }

type TestCommandOther2 struct {
	TestID  string
	Content string
}

func (t *TestCommandOther2) AggregateID() string   { return t.TestID }
func (t *TestCommandOther2) AggregateType() string { return "Test" }
func (t *TestCommandOther2) CommandType() string   { return "TestCommandOther2" }

type MockEventHandler struct {
	events []Event
	recv   chan struct{}
}

func NewMockEventHandler() *MockEventHandler {
	return &MockEventHandler{
		make([]Event, 0),
		make(chan struct{}, 10),
	}
}

func (m *MockEventHandler) HandleEvent(event Event) {
	m.events = append(m.events, event)
	m.recv <- struct{}{}
}

func (m *MockEventHandler) Close() error {
	close(m.recv)
	return nil
}

type MockRepository struct {
	aggregates map[string]Aggregate
}

func (m *MockRepository) Load(aggregateType string, id string) (Aggregate, error) {
	return m.aggregates[id], nil
}

func (m *MockRepository) Save(aggregate Aggregate) error {
	m.aggregates[aggregate.AggregateID()] = aggregate
	return nil
}

func (m *MockRepository) Close() error {
	return nil
}

type MockEventStore struct {
	events []Event
	loaded string
}

func (m *MockEventStore) Save(events []Event) error {
	m.events = append(m.events, events...)
	return nil
}

func (m *MockEventStore) Load(id string) ([]Event, error) {
	m.loaded = id
	return m.events, nil
}

func (m *MockEventStore) Close() error {
	return nil
}

type MockEventBus struct {
	events []Event
}

func (m *MockEventBus) PublishEvent(event Event) {
	m.events = append(m.events, event)
}

func (m *MockEventBus) AddHandler(handler EventHandler, event Event) {}
func (m *MockEventBus) AddLocalHandler(handler EventHandler)         {}
func (m *MockEventBus) AddGlobalHandler(handler EventHandler)        {}
