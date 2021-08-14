package gocache

import (
	"encoding/json"
	"sync"
)

type BoolCache struct {
	m *sync.RWMutex
	v map[string]bool
}

//Get a value
func (s *BoolCache) Get(k string) (v bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v[k]
}

//Set a value
func (s *BoolCache) Set(k string, v bool) {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = v

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
		s = &BoolCache{m: new(sync.RWMutex), v: make(map[string]bool)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &BoolCache{m: new(sync.RWMutex), v: make(map[string]bool)}
}

func (s *BoolCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *BoolCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}
