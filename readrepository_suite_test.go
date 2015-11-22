package eventhorizon

import (
	"time"

	. "gopkg.in/check.v1"
)

type ReadRepositorySuite struct {
	repo ReadRepository
}

func (s *ReadRepositorySuite) Setup(repo ReadRepository) {
	s.repo = repo
}

func (s *ReadRepositorySuite) TestSave(c *C) {
	// Simple save.
	repo := s.repo
	model := NewTestModel("model1")
	repo.Save(model.ID, model)
	all, err := repo.FindAll()
	c.Assert(err, Equals, nil)
	c.Assert(len(all), Equals, 1)
	c.Assert(all[0], DeepEquals, model)
}

func (s *ReadRepositorySuite) TestSaveOverwrite(c *C) {
	// Overwrite with same ID.
	repo := s.repo
	model := NewTestModel("model1")
	repo.Save(model.ID, model)
	model.Content = "model2"
	repo.Save(model.ID, model)
	all, err := repo.FindAll()
	c.Assert(err, Equals, nil)
	c.Assert(len(all), Equals, 1)
	c.Assert(all[0], DeepEquals, model)
}

func (s *ReadRepositorySuite) TestFind(c *C) {
	repo := s.repo
	model := NewTestModel("model1")
	err := repo.Save(model.ID, model)
	c.Assert(err, Equals, nil)
	result, err := repo.Find(model.ID)
	c.Assert(err, Equals, nil)
	c.Assert(result, DeepEquals, model)
}

func (s *ReadRepositorySuite) TestFindEmptyRepo(c *C) {
	repo := s.repo
	result, err := repo.Find(NewUUID())
	c.Assert(err, ErrorMatches, "could not find model")
	c.Assert(result, Equals, nil)
}

func (s *ReadRepositorySuite) TestFindNonExistingID(c *C) {
	repo := s.repo
	err := repo.Save(NewUUID(), NewTestModel("model1"))
	c.Assert(err, Equals, nil)
	result, err := repo.Find(NewUUID())
	c.Assert(err, ErrorMatches, "could not find model")
	c.Assert(result, Equals, nil)
}

func (s *ReadRepositorySuite) TestFindAllOne(c *C) {
	// Find one.
	repo := s.repo
	model := NewTestModel("model1")
	err := repo.Save(model.ID, model)
	c.Assert(err, Equals, nil)
	result, err := repo.FindAll()
	c.Assert(err, Equals, nil)
	c.Assert(result, DeepEquals, []interface{}{model})
}

func (s *ReadRepositorySuite) TestFindAllTwo(c *C) {
	// Find two.
	repo := s.repo
	model := NewTestModel("model1")
	err := repo.Save(model.ID, model)
	c.Assert(err, Equals, nil)
	model2 := NewTestModel("model2")
	err = repo.Save(model2.ID, model2)
	c.Assert(err, Equals, nil)

	result, err := repo.FindAll()
	c.Assert(err, Equals, nil)
	var sum string
	for _, v := range result {
		sum += v.(*TestModel).Content
	}
	c.Assert(sum, Equals, "model1model2")
}

func (s *ReadRepositorySuite) TestFindAllNone(c *C) {
	// Find none.
	repo := s.repo
	result, err := repo.FindAll()
	c.Assert(err, Equals, nil)
	c.Assert(result, DeepEquals, []interface{}{})
}

func (s *ReadRepositorySuite) TestRemove(c *C) {
	// Simple remove.
	repo := s.repo
	model := NewTestModel("model1")
	id := model.ID
	repo.Save(id, model)
	err := repo.Remove(id)
	c.Assert(err, Equals, nil)
	result, err := repo.FindAll()
	c.Assert(err, Equals, nil)
	c.Assert(len(result), Equals, 0)
}

func (s *ReadRepositorySuite) TestRemoveNonExistingID(c *C) {
	// Non existing ID.
	repo := s.repo
	id := NewUUID()
	repo.Save(id, NewTestModel("content"))
	err := repo.Remove(NewUUID())
	c.Assert(err, ErrorMatches, "could not find model")
	result, err := repo.FindAll()
	c.Assert(err, Equals, nil)
	c.Assert(len(result), Equals, 1)
}

type TestModel struct {
	ID        UUID      `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func NewTestModel(content string) *TestModel {
	return NewTestModelWithID(NewUUID(), content)
}

func NewTestModelWithID(id UUID, content string) *TestModel {
	return &TestModel{id, content, time.Now().Round(time.Millisecond)}
}
