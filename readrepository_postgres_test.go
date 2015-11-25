// +build postgres

package eventhorizon

import . "gopkg.in/check.v1"

var _ = Suite(&PostgresReadRepositorySuite{})

type PostgresReadRepositorySuite struct {
	url string
	RemoteReadRepositorySuite
}

func (s *PostgresReadRepositorySuite) SetUpSuite(c *C) {
	s.url = initializePostgresURL()
}

func (s *PostgresReadRepositorySuite) SetUpTest(c *C) {
	repo, err := NewPostgresReadRepository(s.url, "testmodel")
	c.Assert(err, IsNil)
	repo.SetModel(func() interface{} { return &TestModel{} })
	repo.Clear()

	s.Setup(repo)
}

func (s *PostgresReadRepositorySuite) Test_NewPostgresReadRepository(c *C) {
	repo, err := NewPostgresReadRepository(s.url, "testmodel")
	c.Assert(repo, NotNil)
	c.Assert(err, IsNil)
}
