package gocache

import (
	"encoding/json"
	"sync"
)

type FloatCache struct {
	m *sync.RWMutex
	v map[string]float64
}

//Get a value
func (s *FloatCache) Get(k string) (v float64) {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v[k]
}

//Set a value
func (s *FloatCache) Set(k string, v float64) {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = v

}

func (s *FloatCache) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

//Delete a value
func (s *FloatCache) Delete(k string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

}

func (s *FloatCache) Add(k string, v float64) float64 {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] += v
	return s.v[k]
}

func (s *FloatCache) Mul(k string, v float64) float64 {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] *= v
	return s.v[k]
}

func (s *FloatCache) Div(k string, v float64) float64 {
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] /= v
	return s.v[k]
}

func (s *FloatCache) GetKeys() (out []string) {
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

func NewFloatCache(m ...map[string]float64) (s *FloatCache) {
	if len(m) > 0 {
		s = &FloatCache{m: new(sync.RWMutex), v: make(map[string]float64)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &FloatCache{m: new(sync.RWMutex), v: make(map[string]float64)}
}

func (s *FloatCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *FloatCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}
