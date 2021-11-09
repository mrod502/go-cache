package gocache

type DB interface {
	Get(string) ([]byte, error)
	Put(string, Object) error
	Exists(string) (bool, error)
	Delete(string) error
	Where(Matcher) ([]Object, error)
	Keys() []string
}
