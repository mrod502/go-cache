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
}

//Get a value
func (s *InterfaceCache) Get(k string) interface{} {
	s.m.RLock()
	defer s.m.RUnlock()
	if v, ok := s.v[k]; ok {
		return v.load()
	}
	return nil
}

//Set a value
func (s *InterfaceCache) Set(k string, v interface{}) {
	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.store(v)
		s.m.RUnlock()
		return
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = s.newContainer(v)
}

func (s *InterfaceCache) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

//Delete a value
func (s *InterfaceCache) Delete(k string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

}

func (s *InterfaceCache) GetKeys() (out []string) {
	s.m.RLock()
	defer s.m.RUnlock()
	out = make([]string, len(s.v))
	var i int
	for k := range s.v {
		out[i] = k
		i++
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

func (s *InterfaceCache) Query(q Qry, a QueryAppender) {
	for _, v := range s.GetKeys() {
		if q.Match(s.Get(v)) {
			if a(q) {
				return
			}
		}
	}

}

type QueryAppender func(v interface{}) bool

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
		err = s.db.Put(k, s.v[k].load())
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
