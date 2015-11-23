// +build postgres

package eventhorizon

import (
	"fmt"
	"os"

	. "gopkg.in/check.v1"
)

var _ = Suite(&PostgresEventStoreSuite{})

type PostgresEventStoreSuite struct {
	url string
	RemoteEventStoreSuite
}

func initializePostgresURL() string {
	url := os.Getenv("POSTGRES_URL")
	defer func() {
		fmt.Println("using url:", url)
	}()
	if url == "" {
		url = os.Getenv("WERCKER_POSTGRESQL_URL")
	}
	if url != "" {
		return url
	}

	user := os.Getenv("POSTGRES_USERNAME")
	if user == "" {
		user = "postgres"
	}

	database := os.Getenv("POSTGRES_DATABASE")
	if database == "" {
		database = "postgres"
	}

	password := os.Getenv("POSTGRES_ENV_POSTGRES_PASSWORD")
	if password != "" {
		password = " password=" + password
	}

	url = fmt.Sprintf("host=%s port=%s user=%s %s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_PORT_5432_TCP_ADDR"), os.Getenv("POSTGRES_PORT_5432_TCP_PORT"),
		user, password, database)
	return url
}

func (s *PostgresEventStoreSuite) SetUpSuite(c *C) {
	s.url = initializePostgresURL()
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
