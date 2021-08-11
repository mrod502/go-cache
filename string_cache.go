package gocache

import (
	"errors"
	"sync"
)

var (
	ErrKey = errors.New("key not found")
)

type StringCache struct {
	m *sync.RWMutex
	v map[string]string
}

//Get a value
func (s *StringCache) Get(k string) (v string) {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v[k]
}

//Set a value
func (s *StringCache) Set(k string, v string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = v

}

//Delete a value
func (s *StringCache) Delete(k string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

}

func (s *StringCache) GetKeys() (out []string) {
	s.m.RLock()
	defer s.m.RUnlock()
	out = make([]string, len(s.v))
	var i int
	for k := range s.v {
		out[i] = k
	}
	return out
}

func NewStringCache(m ...map[string]string) (s *StringCache) {
	if len(m) > 0 {
		s = &StringCache{m: new(sync.RWMutex), v: make(map[string]string)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &StringCache{m: new(sync.RWMutex), v: make(map[string]string)}
}
