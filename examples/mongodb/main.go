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

// Package example contains a simple runnable example of a CQRS/ES app.

// +build mongo

package main

import (
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/examples/common"
)

func main() {
	eventBus := eventhorizon.NewInternalEventBus()
	commandBus := eventhorizon.NewInternalCommandBus()
	addr := "localhost"
	db := "demo"

	newEventStore := func() (eventhorizon.EventStore, error) {
		return eventhorizon.NewMongoEventStore(eventBus, addr, db)
	}

	newReadRepository := func(name string) (eventhorizon.ReadRepository, error) {
		return eventhorizon.NewMongoReadRepository(addr, db, name)
	}

	common.Run(eventBus, commandBus, newEventStore, newReadRepository)
}
