package gocache

import (
	"encoding/json"
	"sync"

	"go.uber.org/atomic"
	"gopkg.in/yaml.v3"
)

type BoolCache struct {
	m *sync.RWMutex
	v map[string]*atomic.Bool
}

//Get a value
func (s *BoolCache) Get(k string) (v bool) {
	s.m.RLock()
	defer s.m.RUnlock()
	if v, ok := s.v[k]; ok {
		return v.Load()
	}
	return false
}

//Set a value
func (s *BoolCache) Set(k string, v bool) {
	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.Store(v)
		s.m.RUnlock()
		return
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = atomic.NewBool(v)
}

func (s *BoolCache) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

//Delete a value
func (s *BoolCache) Delete(k string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)
}

func (s *BoolCache) GetKeys() (out []string) {
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

func NewBoolCache(m ...map[string]bool) (s *BoolCache) {
	if len(m) > 0 {
		s = &BoolCache{m: new(sync.RWMutex), v: make(map[string]*atomic.Bool)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &BoolCache{m: new(sync.RWMutex), v: make(map[string]*atomic.Bool)}
}

func (s *BoolCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *BoolCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

func (s *BoolCache) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &s.v)
}

func (s *BoolCache) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.v)
}
