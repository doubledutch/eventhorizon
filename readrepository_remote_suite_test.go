package eventhorizon

import . "gopkg.in/check.v1"

type RemoteReadRepositorySuite struct {
	repo RemoteReadRepository
	ReadRepositorySuite
}

func (s *RemoteReadRepositorySuite) Setup(repo RemoteReadRepository) {
	s.repo = repo
	s.ReadRepositorySuite.Setup(repo)
}

func (s *RemoteReadRepositorySuite) TearDownTest(c *C) {
	s.repo.Close()
}

func (s *RemoteReadRepositorySuite) Test_SaveFind(c *C) {
	model1 := NewTestModel("model1")
	err := s.repo.Save(model1.ID, model1)
	c.Assert(err, IsNil)
	model, err := s.repo.Find(model1.ID)
	c.Assert(err, IsNil)
	c.Assert(model, DeepEquals, model1)
}

func (s *RemoteReadRepositorySuite) Test_FindAll(c *C) {
	model1 := NewTestModel("model1")
	model2 := NewTestModel("model2")
	err := s.repo.Save(model1.ID, model1)
	c.Assert(err, IsNil)
	err = s.repo.Save(model2.ID, model2)
	c.Assert(err, IsNil)
	models, err := s.repo.FindAll()
	c.Assert(err, IsNil)
	c.Assert(models, HasLen, 2)
}

func (s *RemoteReadRepositorySuite) Test_Remove(c *C) {
	model1 := NewTestModel("model1")
	err := s.repo.Save(model1.ID, model1)
	c.Assert(err, IsNil)
	model, err := s.repo.Find(model1.ID)
	c.Assert(err, IsNil)
	c.Assert(model, NotNil)
	err = s.repo.Remove(model1.ID)
	c.Assert(err, IsNil)
	model, err = s.repo.Find(model1.ID)
	c.Assert(err, Equals, ErrModelNotFound)
	c.Assert(model, IsNil)
}
