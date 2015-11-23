// +build rabbitmq

package eventhorizon

import (
	"fmt"
	"os"

	. "gopkg.in/check.v1"
)

var _ = Suite(&RabbitMQCommandBusSuite{})

type RabbitMQCommandBusSuite struct {
	uri string
	CommandBusSuite
	rbus *RabbitMQCommandBus
}

func rabbitmqURI() string {
	var creds string
	user := os.Getenv("RABBITMQ_ENV_RABBITMQ_DEFAULT_USER")
	pass := os.Getenv("RABBITMQ_ENV_RABBITMQ_DEFAULT_PASSWORD")
	if user != "" && pass != "" {
		creds = fmt.Sprintf("%s:%s@", user, pass)
	}

	host := os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR")
	if host == "" {
		host = os.Getenv("WERCKER_RABBITMQ_HOST")
	}
	port := os.Getenv("RABBITMQ_PORT_5672_TCP_PORT")
	if port == "" {
		port = os.Getenv("WERCKER_RABBITMQ_PORT")
	}
	if host != "" && port != "" {
		uri := fmt.Sprintf("amqp://%s%s:%s/", creds, host, port)
		return uri
	}
	uri := os.Getenv("RABBITMQ_URI")

	if uri == "" {
		uri = "amqp://localhost:5672/"
	}
	return uri
}

func (s *RabbitMQCommandBusSuite) SetUpSuite(c *C) {
	s.uri = rabbitmqURI()
}

func (s *RabbitMQCommandBusSuite) SetUpTest(c *C) {
	var err error
	s.rbus, err = NewRabbitMQCommandBus(s.uri, "test", "test")
	c.Assert(err, Equals, nil)
	c.Assert(s.rbus, Not(Equals), nil)
	s.rbus.RegisterCommandType(&TestCommand{}, func() Command { return &TestCommand{} })
	c.Assert(err, Equals, nil)
	s.Setup(s.rbus)
}

func (s *RabbitMQCommandBusSuite) TearDownTest(c *C) {
	err := s.rbus.Close()
	c.Assert(err, Equals, nil)
}

func (s *RabbitMQCommandBusSuite) Test_NewHandlerCommandBus(c *C) {
	bus, err := NewRabbitMQCommandBus(s.uri, "test", "test")
	c.Assert(err, Equals, nil)
	c.Assert(bus, Not(Equals), nil)
}
