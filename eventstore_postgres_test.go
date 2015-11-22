package eventhorizon

import (
	"os"

	. "gopkg.in/check.v1"
)

var _ = Suite(&PostgresEventStoreSuite{})

type PostgresEventStoreSuite struct {
	url string
	RemoteEventStoreSuite
}

func (s *PostgresEventStoreSuite) SetUpSuite(c *C) {
	s.url = os.Getenv("WERCKER_POSTGRESQL_URL")
}

func (s *PostgresEventStoreSuite) SetUpTest(c *C) {
	bus := &MockEventBus{
		events: make([]Event, 0),
	}
	store, err := NewPostgresEventStore(bus, s.url)
	c.Assert(err, IsNil)

	s.RemoteEventStoreSuite.Setup(store, c)
}

func (s *PostgresEventStoreSuite) Test_NewPostgresEventStore(c *C) {
	bus := &MockEventBus{
		events: make([]Event, 0),
	}
	store, err := NewPostgresEventStore(bus, s.url)
	c.Assert(store, NotNil)
	c.Assert(err, IsNil)
}
