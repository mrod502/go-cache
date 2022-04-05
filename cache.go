package gocache

import (
	"encoding/json"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type container[T any] struct {
	m        *sync.RWMutex
	v        T
	deadline time.Time
}

func (c *container[T]) load() T {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.v
}
func (c *container[T]) store(v T) {
	c.m.Lock()
	defer c.m.Unlock()
	c.v = v
}

type Stringer interface {
	String() string
}

type Cache[T any] struct {
	m             *sync.RWMutex
	v             map[string]*container[T]
	expire        time.Duration
	sleepInterval time.Duration
}

//Get a value --
func (s *Cache[T]) Get(k string) (interface{}, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	if v, ok := s.v[k]; ok {
		return v.load(), nil
	}
	return nil, ErrKey
}

func (s *Cache[T]) Where(m func(interface{}) bool) (v []interface{}, err error) {
	return s.memQuery(m)
}

func (s *Cache[T]) memQuery(m func(interface{}) bool) (v []interface{}, err error) {
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
func (s *Cache[T]) Set(k string, v T) error {
	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.store(v)
		s.m.RUnlock()
		return nil
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = s.newContainer(v)
	return nil
}

func (s *Cache[T]) Exists(k string) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

//Delete a value
func (s *Cache[T]) Delete(k string) (err error) {

	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

	return nil
}

func (s *Cache[T]) GetKeys(fromDb ...bool) (out []string) {
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

func New[T any](m ...map[string]T) (s *Cache[T]) {
	if len(m) > 0 {
		s = &Cache[T]{m: new(sync.RWMutex), v: make(map[string]*container[T])}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &Cache[T]{m: new(sync.RWMutex), v: make(map[string]*container[T])}
}

func (s *Cache[T]) janitor() {
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

func (s *Cache[T]) WithExpiration(e time.Duration) *Cache[T] {
	s.expire = e
	if e > (30 * time.Second) {
		s.sleepInterval = e
	} else {
		s.sleepInterval = 30 * time.Second
	}
	go s.janitor()
	return s
}

func (c *Cache[T]) newContainer(v T) *container[T] {
	return &container[T]{m: &sync.RWMutex{}, v: v, deadline: time.Now().Add(c.expire)}
}

func (s *Cache[T]) unCache(k string) (err error) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)
	return
}

func (s *Cache[T]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *Cache[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

func (s *Cache[T]) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &s.v)
}

func (s *Cache[T]) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.v)
}
