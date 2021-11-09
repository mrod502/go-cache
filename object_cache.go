package gocache

import (
	"encoding/json"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type ObjectCache struct {
	m             *sync.RWMutex
	v             map[string]*container
	db            DB
	persist       bool
	expire        time.Duration
	sleepInterval time.Duration
	writeQ        chan action
	waitForRes    bool
}

type Object interface {
	Create() error
	Destroy() error
}

func (s *ObjectCache) Get(k string) (Object, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	if v, ok := s.v[k]; ok {
		return v.load(), nil
	}
	r := <-s.aGet(k)
	return r.res, r.err
}

func (s *ObjectCache) Where(m Matcher) (v []Object, err error) {
	if !s.persist {
		return s.memQuery(m)
	}
	return s.dbQuery(m)
}

func (s *ObjectCache) memQuery(m Matcher) (v []Object, err error) {
	s.m.RLock()
	defer s.m.RUnlock()
	for _, k := range s.GetKeys() {
		val := s.v[k].load()
		if m(val) {
			v = append(v, val)
		}
	}
	return
}

func (s *ObjectCache) dbQuery(m Matcher) (v []Object, err error) {
	return s.db.Where(m)
}

//Set a value
func (s *ObjectCache) Set(k string, v Object) error {
	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.store(v)
		s.m.RUnlock()
		return nil
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = s.newContainer(v)

	if s.persist {
		res := s.aPut(k, v)
		if s.waitForRes {
			return (<-res).err
		}
	}
	return nil
}

func (s *ObjectCache) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	if ok || !s.persist {
		return ok
	}
	return (<-s.aExists(k)).exist
}

//Delete a value
func (s *ObjectCache) Delete(k string) (err error) {
	v, err := s.Get(k)
	if err != nil {
		return err
	}
	defer v.Destroy()
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)
	if s.persist {
		res := s.aDelete(k)
		if s.waitForRes {
			return (<-res).err
		}
	}
	return nil
}

func (s *ObjectCache) GetKeys(fromDb ...bool) (out []string) {
	s.m.RLock()
	defer s.m.RUnlock()
	out = make([]string, len(s.v))
	var i int
	for k := range s.v {
		out[i] = k
		i++
	}
	if s.persist {
		if Or(fromDb...) {
			out = append(out, s.db.Keys()...)
		}
	}
	return out
}

func NewObjectCache(m ...map[string]Object) (s *ObjectCache) {
	if len(m) > 0 {
		s = &ObjectCache{m: new(sync.RWMutex), v: make(map[string]*container)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &ObjectCache{m: new(sync.RWMutex), v: make(map[string]*container)}
}

func (s *ObjectCache) janitor() {
	for {
		time.Sleep(s.sleepInterval)
		now := time.Now()
		for _, k := range s.GetKeys() {
			s.m.RLock()
			if v, ok := s.v[k]; ok {
				if v.deadline.Before(now) {
					s.m.RUnlock()
					s.unCache(k)
				} else {
					s.m.RUnlock()
				}
			} else {
				s.m.RUnlock()
			}
		}
	}
}

func (s *ObjectCache) WithDb(d DB) *ObjectCache {
	s.persist = true
	s.db = d
	go s.writer()
	return s
}

func (s *ObjectCache) writer() {
	var res actionResponse
	for {
		action := <-s.writeQ
		switch action.act {
		case actionGet:

		case actionPut:
			res.err = s.db.Put(action.k, action.v)
		case actionExist:
			res.exist, res.err = s.db.Exists(action.k)
		case actionDelete:
			res.err = s.db.Delete(action.k)
		case actionQuery:
			res.qRes, res.err = s.db.Where(action.qry)
		}
		if action.wantRes {
			action.resChan <- res
			close(action.resChan)
		}
	}
}

func (s *ObjectCache) WithExpiration(e time.Duration) *ObjectCache {
	s.expire = e
	if e > (30 * time.Second) {
		s.sleepInterval = e
	} else {
		s.sleepInterval = 30 * time.Second
	}
	go s.janitor()
	return s
}

func (s *ObjectCache) unCache(k string) (err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.persist {
		res := s.aPut(k, s.v[k].load())
		if s.waitForRes {
			err = (<-res).err
		}
	}
	delete(s.v, k)
	return
}

func (s *ObjectCache) DispatchEvent(e func(Object) error) error {
	var err error
	s.m.Lock()
	defer s.m.Unlock()
	for _, v := range s.v {
		if er := e(v.load()); er != nil {
			err = er
		}
	}
	return err
}

func (s *ObjectCache) aGet(k string) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionGet,
		k:       k,
		wantRes: true,
		resChan: ch,
	}
	return ch
}

func (s *ObjectCache) aPut(k string, v Object) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionPut,
		k:       k,
		v:       v,
		wantRes: s.waitForRes,
		resChan: ch,
	}
	return ch
}

func (s *ObjectCache) aDelete(k string) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionDelete,
		k:       k,
		wantRes: s.waitForRes,
		resChan: ch,
	}
	return ch
}

func (s *ObjectCache) aExists(k string) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionExist,
		k:       k,
		wantRes: true,
		resChan: ch,
	}
	return ch
}

func (s *ObjectCache) aQuery(q Matcher) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionQuery,
		qry:     q,
		wantRes: s.waitForRes,
		resChan: ch,
	}
	return ch
}

func (s *ObjectCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *ObjectCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

func (s *ObjectCache) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &s.v)
}

func (s *ObjectCache) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.v)
}
