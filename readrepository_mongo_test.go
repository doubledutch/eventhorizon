// Copyright (c) 2015 - Max Persson <max@looplab.se>
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

// +build mongo

package eventhorizon

import (
	"os"
	"time"

	"github.com/odeke-em/go-uuid"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	. "gopkg.in/check.v1"
)

var _ = Suite(&MongoReadRepositorySuite{})

type MongoReadRepositorySuite struct {
	url  string
	repo *MongoReadRepository
	RemoteReadRepositorySuite
}

func (s *MongoReadRepositorySuite) SetUpSuite(c *C) {
	// Support Wercker testing with MongoDB.
	host := os.Getenv("MONGO_PORT_27017_TCP_ADDR")
	port := os.Getenv("MONGO_PORT_27017_TCP_PORT")

	if host != "" && port != "" {
		s.url = host + ":" + port
	} else {
		s.url = "localhost"
	}
}

func (s *MongoReadRepositorySuite) SetUpTest(c *C) {
	var err error
	s.repo, err = NewMongoReadRepository(s.url, "test", "testmodel")
	c.Assert(err, IsNil)
	s.repo.SetModel(func() interface{} { return &TestModel{} })
	s.repo.Clear()

	s.Setup(s.repo)
}

func (s *MongoReadRepositorySuite) TearDownTest(c *C) {
	s.repo.Close()
}

func (s *MongoReadRepositorySuite) Test_NewMongoReadRepository(c *C) {
	repo, err := NewMongoReadRepository(s.url, "test", "testmodel")
	c.Assert(repo, NotNil)
	c.Assert(err, IsNil)
}

func (s *MongoReadRepositorySuite) Test_FindCustom(c *C) {
	model1 := &TestModel{uuid.New(), "model1", time.Now().Round(time.Millisecond)}
	err := s.repo.Save(model1.ID, model1)
	c.Assert(err, IsNil)
	models, err := s.repo.FindCustom(func(c *mgo.Collection) *mgo.Query {
		return c.Find(bson.M{"content": "model1"})
	})
	c.Assert(err, IsNil)
	c.Assert(models, HasLen, 1)
	c.Assert(models[0], DeepEquals, model1)
}
