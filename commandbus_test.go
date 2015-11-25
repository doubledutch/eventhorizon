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
	. "gopkg.in/check.v1"
)

var _ = Suite(&InternalCommandBusSuite{})

type InternalCommandBusSuite struct {
	CommandBusSuite
}

func (s *InternalCommandBusSuite) SetUpTest(c *C) {
	bus := NewInternalCommandBus()
	s.Setup(bus)
}

func (s *InternalCommandBusSuite) Test_NewHandlerCommandBus(c *C) {
	bus := NewInternalCommandBus()
	c.Assert(bus, Not(Equals), nil)
}
