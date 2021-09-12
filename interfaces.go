package gocache

type Qry interface {
	Match(interface{}) bool
}

type DB interface {
	Get(string) (interface{}, error)
	Put(string, interface{}) error
	Exists(string) (bool, error)
	Delete(string) error
	Where(Matcher) (interface{}, error)
	Keys() []string
}

type Dispatcher func(interface{}) error
