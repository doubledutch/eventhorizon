// +build mongo

package eventhorizon

import (
	"os"

	. "gopkg.in/check.v1"
)

var _ = Suite(&MongoEventStoreSuite{})

type MongoEventStoreSuite struct {
	url string
	RemoteEventStoreSuite
}

func (s *MongoEventStoreSuite) SetUpSuite(c *C) {
	// Support Wercker testing with MongoDB.
	host := os.Getenv("MONGO_PORT_27017_TCP_ADDR")
	if host == "" {
		host = os.Getenv("WERCKER_MONGODB_HOST")
	}
	port := os.Getenv("MONGO_PORT_27017_TCP_PORT")
	if port == "" {
		port = os.Getenv("WERCKER_MONGODB_PORT")
	}

	if host != "" && port != "" {
		s.url = host + ":" + port
	} else {
		s.url = "localhost"
	}
}

func (s *MongoEventStoreSuite) SetUpTest(c *C) {
	bus := &MockEventBus{
		events: make([]Event, 0),
	}
	store, err := NewMongoEventStore(bus, s.url, "test")
	c.Assert(err, IsNil)

	s.RemoteEventStoreSuite.Setup(store, c)
}

func (s *MongoEventStoreSuite) Test_NewMongoEventStore(c *C) {
	bus := &MockEventBus{
		events: make([]Event, 0),
	}
	store, err := NewMongoEventStore(bus, s.url, "test")
	c.Assert(store, NotNil)
	c.Assert(err, IsNil)
}
