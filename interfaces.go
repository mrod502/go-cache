package gocache

type Qry interface {
	Match(interface{}) bool
}

type Matcher interface {
	Match(Qry) bool
}
