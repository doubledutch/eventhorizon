package eventhorizon

// RemoteReadRepository is a read repository that uses a networked service
type RemoteReadRepository interface {
	ReadRepository
	SetModel(factory func() interface{})
	Close() error
	Clear() error
}
