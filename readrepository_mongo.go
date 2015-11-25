package eventhorizon

import "gopkg.in/mgo.v2"

// MongoReadRepository implements an MongoDB repository of read models.
type MongoReadRepository struct {
	session    *mgo.Session
	db         string
	collection string
	factory    func() interface{}
}

// NewMongoReadRepository creates a new MongoReadRepository.
func NewMongoReadRepository(url, database, collection string) (*MongoReadRepository, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, ErrCouldNotDialDB
	}

	session.SetMode(mgo.Strong, true)
	session.SetSafe(&mgo.Safe{W: 1})

	return NewMongoReadRepositoryWithSession(session, database, collection)
}

// NewMongoReadRepositoryWithSession creates a new MongoReadRepository with a session.
func NewMongoReadRepositoryWithSession(session *mgo.Session, database, collection string) (*MongoReadRepository, error) {
	if session == nil {
		return nil, ErrNoDBSession
	}

	r := &MongoReadRepository{
		session:    session,
		db:         database,
		collection: collection,
	}

	return r, nil
}

// Save saves a read model with id to the repository.
func (r *MongoReadRepository) Save(id string, model interface{}) error {
	sess := r.session.Copy()
	defer sess.Close()

	if _, err := sess.DB(r.db).C(r.collection).UpsertId(id, model); err != nil {
		return ErrCouldNotSaveModel
	}
	return nil
}

// Find returns one read model with using an id. Returns
// ErrModelNotFound if no model could be found.
func (r *MongoReadRepository) Find(id string) (interface{}, error) {
	sess := r.session.Copy()
	defer sess.Close()

	if r.factory == nil {
		return nil, ErrModelNotSet
	}

	model := r.factory()
	err := sess.DB(r.db).C(r.collection).FindId(id).One(model)
	if err != nil {
		return nil, ErrModelNotFound
	}

	return model, nil
}

// FindCustom uses a callback to specify a custom query.
func (r *MongoReadRepository) FindCustom(callback func(*mgo.Collection) *mgo.Query) ([]interface{}, error) {
	sess := r.session.Copy()
	defer sess.Close()

	if r.factory == nil {
		return nil, ErrModelNotSet
	}

	collection := sess.DB(r.db).C(r.collection)
	query := callback(collection)

	iter := query.Iter()
	result := []interface{}{}
	model := r.factory()
	for iter.Next(model) {
		result = append(result, model)
		model = r.factory()
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return result, nil
}

// FindAll returns all read models in the repository.
func (r *MongoReadRepository) FindAll() ([]interface{}, error) {
	sess := r.session.Copy()
	defer sess.Close()

	if r.factory == nil {
		return nil, ErrModelNotSet
	}

	iter := sess.DB(r.db).C(r.collection).Find(nil).Iter()
	result := []interface{}{}
	model := r.factory()
	for iter.Next(model) {
		result = append(result, model)
		model = r.factory()
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return result, nil
}

// Remove removes a read model with id from the repository. Returns
// ErrModelNotFound if no model could be found.
func (r *MongoReadRepository) Remove(id string) error {
	sess := r.session.Copy()
	defer sess.Close()

	err := sess.DB(r.db).C(r.collection).RemoveId(id)
	if err != nil {
		return ErrModelNotFound
	}

	return nil
}

// SetModel sets a factory function that creates concrete model types.
func (r *MongoReadRepository) SetModel(factory func() interface{}) {
	r.factory = factory
}

// SetDB sets the database session and database.
func (r *MongoReadRepository) SetDB(db string) {
	r.db = db
}

// Clear clears the read model database.
func (r *MongoReadRepository) Clear() error {
	if err := r.session.DB(r.db).C(r.collection).DropCollection(); err != nil {
		return ErrCouldNotClearDB
	}
	return nil
}

// Close closes a database session.
func (r *MongoReadRepository) Close() error {
	r.session.Close()
	return nil
}
