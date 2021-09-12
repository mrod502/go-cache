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
	v       interface{}
	qry     Matcher
	wantRes bool
	resChan chan actionResponse
}

type actionResponse struct {
	err error
	res interface{}
}
