package gocache

type actionType byte

const (
	actionGet actionType = iota
	actionPut
	actionDelete
	actionExist
	actionQuery
)

type action struct {
	act     actionType
	k       string
	v       Object
	qry     Matcher
	wantRes bool
	resChan chan actionResponse
}

type actionResponse struct {
	err   error
	res   Object
	qRes  []Object
	exist bool
}
