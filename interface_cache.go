package gocache

import (
	"encoding/json"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type InterfaceCache struct {
	m             *sync.RWMutex
	v             map[string]*container
	db            DB
	persist       bool
	expire        time.Duration
	sleepInterval time.Duration
	writeQ        chan action
	waitForRes    bool
}

//Get a value
func (s *InterfaceCache) Get(k string) interface{} {
	s.m.RLock()
	defer s.m.RUnlock()
	if v, ok := s.v[k]; ok {
		return v.load()
	}
	return (<-s.aGet(k)).res
}

//Set a value
func (s *InterfaceCache) Set(k string, v interface{}) error {
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

func (s *InterfaceCache) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	if ok || !s.persist {
		return ok
	}
	return (<-s.aExists(k)).res.(bool)
}

//Delete a value
func (s *InterfaceCache) Delete(k string) (err error) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)
	if s.persist {
		res := s.aDelete(k)
		if s.waitForRes {
			err = (<-res).err
		}
	}
	return
}

func (s *InterfaceCache) GetKeys(fromDb ...bool) (out []string) {
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

func NewInterfaceCache(m ...map[string]interface{}) (s *InterfaceCache) {
	if len(m) > 0 {
		s = &InterfaceCache{m: new(sync.RWMutex), v: make(map[string]*container)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &InterfaceCache{m: new(sync.RWMutex), v: make(map[string]*container)}
}

func (s *InterfaceCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *InterfaceCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

func (s *InterfaceCache) Query(q Matcher) error {
	for _, v := range s.GetKeys() {
		q(v)
	}
	if s.persist {
		res := s.aQuery(q)
		if s.waitForRes {
			return (<-res).err
		}
	}
	return nil
}

func (s *InterfaceCache) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &s.v)
}

func (s *InterfaceCache) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.v)
}

func (s *InterfaceCache) janitor() {
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

func (s *InterfaceCache) WithDb(d DB) *InterfaceCache {
	s.persist = true
	s.db = d
	go s.writer()
	return s
}

func (s *InterfaceCache) writer() {
	var res actionResponse
	for {
		action := <-s.writeQ
		switch action.act {
		case actionGet:
			res.res, res.err = s.db.Get(action.k)
		case actionPut:
			res.err = s.db.Put(action.k, action.v)
		case actionExist:
			res.res, res.err = s.db.Exists(action.k)
		case actionDelete:
			res.err = s.db.Delete(action.k)
		case actionQuery:
			res.res, res.err = s.db.Where(action.qry)
		}
		if action.wantRes {
			action.resChan <- res
			close(action.resChan)
		}
	}
}

func (s *InterfaceCache) WithExpiration(e time.Duration) *InterfaceCache {
	s.expire = e
	if e >= 3*time.Second {
		s.sleepInterval = e
	} else {
		s.sleepInterval = 3 * time.Second
	}
	go s.janitor()
	return s
}

func (s *InterfaceCache) unCache(k string) (err error) {
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

func (s *InterfaceCache) DispatchEvent(e func(interface{}) error) error {
	var err error
	s.m.Lock()
	defer s.m.Unlock()
	for _, v := range s.v {
		if er := e(v); er != nil {
			err = er
		}
	}
	return err
}

func (s *InterfaceCache) aGet(k string) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionGet,
		k:       k,
		wantRes: true,
		resChan: ch,
	}
	return ch
}

func (s *InterfaceCache) aPut(k string, v interface{}) chan actionResponse {
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

func (s *InterfaceCache) aDelete(k string) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionDelete,
		k:       k,
		wantRes: s.waitForRes,
		resChan: ch,
	}
	return ch
}

func (s *InterfaceCache) aExists(k string) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionExist,
		k:       k,
		wantRes: true,
		resChan: ch,
	}
	return ch
}

func (s *InterfaceCache) aQuery(q Matcher) chan actionResponse {
	ch := make(chan actionResponse)
	s.writeQ <- action{
		act:     actionQuery,
		qry:     q,
		wantRes: s.waitForRes,
		resChan: ch,
	}
	return ch
}
