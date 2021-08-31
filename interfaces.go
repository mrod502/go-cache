package gocache

type Qry interface {
	Match(interface{}) bool
}

type Matcher interface {
	Match(Qry) bool
}

type DB interface {
	Get(string) (interface{}, error)
	Put(string, interface{}) error
	Exists(string) (bool, error)
	Delete(string) error
}

type Dispatcher func(interface{}) error
