package gocache

import (
	"encoding/json"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type InterfaceCache struct {
	m             *sync.RWMutex
	v             map[string]*icontainer
	expire        time.Duration
	sleepInterval time.Duration
}

//Get a value --
func (s *InterfaceCache) Get(k string) (interface{}, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	if v, ok := s.v[k]; ok {
		return v.load(), nil
	}
	return nil, ErrKey
}

func (s *InterfaceCache) Where(m func(interface{}) bool) (v []interface{}, err error) {
	return s.memQuery(m)
}

func (s *InterfaceCache) memQuery(m func(interface{}) bool) (v []interface{}, err error) {
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
	s.v[k] = s.newiContainer(v)
	return nil
}

func (s *InterfaceCache) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

//Delete a value
func (s *InterfaceCache) Delete(k string) (err error) {

	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

	return nil
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
	return out
}

func NewInterfaceCache(m ...map[string]interface{}) (s *InterfaceCache) {
	if len(m) > 0 {
		s = &InterfaceCache{m: new(sync.RWMutex), v: make(map[string]*icontainer)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &InterfaceCache{m: new(sync.RWMutex), v: make(map[string]*icontainer)}
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
	if e > (30 * time.Second) {
		s.sleepInterval = e
	} else {
		s.sleepInterval = 30 * time.Second
	}
	go s.janitor()
	return s
}

func (s *InterfaceCache) unCache(k string) (err error) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)
	return
}

func (s *InterfaceCache) DispatchEvent(e func(interface{}) error) error {
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

func (s *InterfaceCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *InterfaceCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

func (s *InterfaceCache) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &s.v)
}

func (s *InterfaceCache) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.v)
}
