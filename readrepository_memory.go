package eventhorizon

// MemoryReadRepository implements an in memory repository of read models.
type MemoryReadRepository struct {
	data map[UUID]interface{}
}

// NewMemoryReadRepository creates a new MemoryReadRepository.
func NewMemoryReadRepository() *MemoryReadRepository {
	r := &MemoryReadRepository{
		data: make(map[UUID]interface{}),
	}
	return r
}

// Save saves a read model with id to the repository.
func (r *MemoryReadRepository) Save(id UUID, model interface{}) error {
	r.data[id] = model
	return nil
}

// Find returns one read model with using an id. Returns
// ErrModelNotFound if no model could be found.
func (r *MemoryReadRepository) Find(id UUID) (interface{}, error) {
	if model, ok := r.data[id]; ok {
		return model, nil
	}

	return nil, ErrModelNotFound
}

// FindAll returns all read models in the repository.
func (r *MemoryReadRepository) FindAll() ([]interface{}, error) {
	models := []interface{}{}
	for _, model := range r.data {
		models = append(models, model)
	}
	return models, nil
}

// Remove removes a read model with id from the repository. Returns
// ErrModelNotFound if no model could be found.
func (r *MemoryReadRepository) Remove(id UUID) error {
	if _, ok := r.data[id]; ok {
		delete(r.data, id)
		return nil
	}

	return ErrModelNotFound
}
