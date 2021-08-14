package gocache

import (
	"encoding/json"
	"sync"
)

type IntCache struct {
	m *sync.RWMutex
	v map[string]int
}

//Get a value
func (s *IntCache) Get(k string) (v int) {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v[k]
}

//Set a value
func (s *IntCache) Set(k string, v int) {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = v

}

func (s *IntCache) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

//Delete a value
func (s *IntCache) Delete(k string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

}

func (s *IntCache) Add(k string, v int) int {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] += v
	return s.v[k]
}

func (s *IntCache) Mul(k string, v int) int {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] *= v
	return s.v[k]
}

func (s *IntCache) Div(k string, v int) int {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] /= v
	return s.v[k]
}

func (s *IntCache) GetKeys() (out []string) {
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

func NewIntCache(m ...map[string]int) (s *IntCache) {
	if len(m) > 0 {
		s = &IntCache{m: new(sync.RWMutex), v: make(map[string]int)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &IntCache{m: new(sync.RWMutex), v: make(map[string]int)}
}

func (s *IntCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *IntCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}
