// +build rabbitmq

package eventhorizon

import . "gopkg.in/check.v1"

var _ = Suite(&RabbitMQEventBusSuite{})

type RabbitMQEventBusSuite struct {
	uri string
	EventBusSuite
	bus  RemoteEventBus
	bus2 RemoteEventBus
}

func (s *RabbitMQEventBusSuite) SetUpSuite(c *C) {
	s.uri = rabbitmqURI()
}

func (s *RabbitMQEventBusSuite) SetUpTest(c *C) {
	var err error
	s.bus, err = NewRabbitMQEventBus(s.uri, "test", "bus1")
	c.Assert(err, IsNil)
	c.Assert(s.bus, NotNil)
	err = s.bus.RegisterEventType(&TestEvent{}, func() Event { return &TestEvent{} })
	c.Assert(err, IsNil)

	s.bus2, err = NewRabbitMQEventBus(s.uri, "test", "bus2")
	c.Assert(err, IsNil)
	c.Assert(s.bus2, NotNil)
	err = s.bus2.RegisterEventType(&TestEvent{}, func() Event { return &TestEvent{} })
	c.Assert(err, IsNil)
	s.EventBusSuite.Bus = s.bus
	s.EventBusSuite.Bus2 = s.bus2
}

func (s *RabbitMQEventBusSuite) TearDownTest(c *C) {
	s.bus.Close()
	s.bus2.Close()
}

func (s *RabbitMQEventBusSuite) Test_NewHandlerEventBus(c *C) {
	bus, err := NewRabbitMQEventBus(s.uri, "test", "bus")
	c.Assert(err, IsNil)
	c.Assert(bus, NotNil)
	bus.Close()
}
