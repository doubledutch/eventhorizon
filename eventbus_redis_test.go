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

// +build redis

package eventhorizon

import (
	"os"

	. "gopkg.in/check.v1"
)

var _ = Suite(&RedisEventBusSuite{})

type RedisEventBusSuite struct {
	url string
	EventBusSuite
	bus  RemoteEventBus
	bus2 RemoteEventBus
}

func (s *RedisEventBusSuite) SetUpSuite(c *C) {
	// Support Wercker testing with MongoDB.
	host := os.Getenv("WERCKER_REDIS_HOST")
	port := os.Getenv("WERCKER_REDIS_PORT")

	if host != "" && port != "" {
		s.url = host + ":" + port
	} else {
		s.url = ":6379"
	}
}
func (s *RedisEventBusSuite) SetUpTest(c *C) {
	var err error
	s.bus, err = NewRedisEventBus("test", s.url, "")
	c.Assert(s.bus, NotNil)
	c.Assert(err, IsNil)
	err = s.bus.RegisterEventType(&TestEvent{}, func() Event { return &TestEvent{} })
	c.Assert(err, IsNil)

	s.bus2, err = NewRedisEventBus("test", s.url, "")
	c.Assert(s.bus2, NotNil)
	c.Assert(err, IsNil)
	err = s.bus2.RegisterEventType(&TestEvent{}, func() Event { return &TestEvent{} })
	c.Assert(err, IsNil)

	s.EventBusSuite.Bus = s.bus
	s.EventBusSuite.Bus2 = s.bus2
}

func (s *RedisEventBusSuite) TearDownTest(c *C) {
	s.bus.Close()
	s.bus2.Close()
}

func (s *RedisEventBusSuite) Test_NewHandlerEventBus(c *C) {
	bus, err := NewRedisEventBus("test", s.url, "")
	c.Assert(bus, NotNil)
	c.Assert(err, IsNil)
	bus.Close()
}
