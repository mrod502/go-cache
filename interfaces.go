package gocache

// a db
type DB interface {
	Get(string) (Object, error)
	Put(string, Object) error
	Exists(string) (bool, error)
	Delete(string) error
	Where(Matcher) ([]Object, error)
	Keys() []string
}
