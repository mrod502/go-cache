package gocache

import (
	"time"
)

type DBProps struct {
	Id         []byte    `msgpack:"i"`
	Archived   bool      `msgpack:"d"`
	Created    time.Time `msgpack:"cr"`
	LastEdited time.Time `msgpack:"le"`
}

func (d DBProps) ID() []byte {
	return d.Id
}

func (d *DBProps) SetArchived(a bool) {
	d.Archived = a
}

type DBPropsQuery struct {
	Id         ByteSliceQuery `msgpack:"i,omitempty"`
	Archived   BoolQuery      `msgpack:"d,omitempty"`
	Created    TimeQuery      `msgpack:"cr,omitempty"`
	LastEdited TimeQuery      `msgpack:"le,omitempty"`
}

func (q DBPropsQuery) Match(v DBProps) bool {
	return q.Id.Match(v.Id) && q.Archived.Match(v.Archived) &&
		q.Created.Match(v.Created) && q.LastEdited.Match(v.LastEdited)
}
