package gocache

import (
	"encoding/json"
	"sync"

	"gopkg.in/yaml.v3"
)

type InterfaceCache struct {
	m *sync.RWMutex
	v map[string]interface{}
}

//Get a value
func (s *InterfaceCache) Get(k string) (v interface{}) {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v[k]
}

//Set a value
func (s *InterfaceCache) Set(k string, v interface{}) {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = v

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
		s = &InterfaceCache{m: new(sync.RWMutex), v: make(map[string]interface{})}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &InterfaceCache{m: new(sync.RWMutex), v: make(map[string]interface{})}
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
