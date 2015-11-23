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
	"errors"
)

// ErrHandlerAlreadySet returned when a handler is already registered for a command.
var ErrHandlerAlreadySet = errors.New("handler is already set")

// ErrHandlerNotFound returned when no handler can be found.
var ErrHandlerNotFound = errors.New("no handlers for command")

// CommandHandler is an interface that all handlers of commands should implement.
type CommandHandler interface {
	HandleCommand(Command) error
}

// CommandBus is an interface defining an event bus for distributing events.
type CommandBus interface {
	// PublishCommand publishes a command on the command bus.
	PublishCommand(Command) error
	// HandleCommand handles a command on the command bus.
	HandleCommand(Command) error
	// SetHandler registers a handler with a command.
	SetHandler(CommandHandler, Command) error
}

// RemoteCommandBus is a command bus that using a networked service.
type RemoteCommandBus interface {
	CommandBus
	RegisterCommandType(command Command, factory func() Command) error
	Close() error
}

// InternalCommandBus is a command bus that handles commands with the
// registered CommandHandlers
type InternalCommandBus struct {
	handlers map[string]CommandHandler
}

// NewInternalCommandBus creates a InternalCommandBus.
func NewInternalCommandBus() CommandBus {
	b := &InternalCommandBus{
		handlers: make(map[string]CommandHandler),
	}
	return b
}

// PublishCommand publishes a command to the internal command bus.
func (b *InternalCommandBus) PublishCommand(command Command) error {
	return b.HandleCommand(command)
}

// HandleCommand handles a command with a handler capable of handling it.
func (b *InternalCommandBus) HandleCommand(command Command) error {
	if handler, ok := b.handlers[command.CommandType()]; ok {
		return handler.HandleCommand(command)
	}
	return ErrHandlerNotFound
}

// SetHandler adds a handler for a specific command.
func (b *InternalCommandBus) SetHandler(handler CommandHandler, command Command) error {
	if _, ok := b.handlers[command.CommandType()]; ok {
		return ErrHandlerAlreadySet
	}
	b.handlers[command.CommandType()] = handler
	return nil
}
