package gocache

import (
	"encoding/json"
	"sync"

	"go.uber.org/atomic"
	"gopkg.in/yaml.v3"
)

type FloatCache struct {
	m *sync.RWMutex
	v map[string]*atomic.Float64
}

//Get a value
func (s *FloatCache) Get(k string) (v float64) {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.v[k].Load()
}

//Set a value
func (s *FloatCache) Set(k string, v float64) {

	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.Store(v)
	}
	s.m.RUnlock()
	s.m.Lock()
	s.v[k] = atomic.NewFloat64(v)
	s.m.Unlock()
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
	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.Store(val.Load() + v)
		s.m.RUnlock()
		return val.Load()
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = atomic.NewFloat64(v)

	return s.v[k].Load()
}

func (s *FloatCache) Mul(k string, v float64) float64 {
	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.Store(val.Load() * v)
		s.m.RUnlock()
		return val.Load()
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = atomic.NewFloat64(0)

	return 0
}

func (s *FloatCache) Div(k string, v float64) float64 {
	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.Store(val.Load() / v)
		s.m.RUnlock()
		return val.Load()
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = atomic.NewFloat64(0)

	return 0
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
		s = &FloatCache{m: new(sync.RWMutex), v: make(map[string]*atomic.Float64)}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &FloatCache{m: new(sync.RWMutex), v: make(map[string]*atomic.Float64)}
}

func (s *FloatCache) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *FloatCache) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

func (s *FloatCache) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &s.v)
}

func (s *FloatCache) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.v)
}
