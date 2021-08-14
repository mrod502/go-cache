package gocache

import (
	"encoding/json"
	"sync"
)

type ByteArrayCache struct {
	m *sync.RWMutex
	v map[string][]byte
}

//Get a value
func (s *ByteArrayCache) Get(k string) (v []byte) {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v[k]
}

//Set a value
func (s *ByteArrayCache) Set(k string, v []byte) {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = v

}

func (s *ByteArrayCache) Append(k string, v []byte) []byte {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = append(s.v[k], v...)
	return s.v[k]

}

func (s *ByteArrayCache) Slice(k string, a, b int) []byte {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v[k][a:b]

}

func (s *ByteArrayCache) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

//Delete a value
func (s *ByteArrayCache) Delete(k string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

}

func (s *ByteArrayCache) GetKeys() (out []string) {
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

func NewByteArrayCache(m ...map[string][]byte) (s *ByteArrayCache) {
	if len(m) > 0 {
		s = &ByteArrayCache{m: new(sync.RWMutex), v: make(map[string][]byte)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &ByteArrayCache{m: new(sync.RWMutex), v: make(map[string][]byte)}
}

func (s *ByteArrayCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *ByteArrayCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}
