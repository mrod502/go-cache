package gocache

import (
	"encoding/json"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Container[T any] struct {
	m        *sync.RWMutex
	v        T
	deadline time.Time
}

func NewContainer[T any](v T, deadline time.Time) *Container[T] {
	return &Container[T]{
		m:        &sync.RWMutex{},
		v:        v,
		deadline: deadline,
	}
}

func (c *Container[T]) Load() T {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.v
}
func (c *Container[T]) Store(v T) {
	c.m.Lock()
	defer c.m.Unlock()
	c.v = v
}

func (c *Container[T]) Expired() bool { return time.Now().UnixNano() > c.deadline.UnixNano() }

type Stringer interface {
	String() string
}

type Cache[T any, K comparable] struct {
	m             *sync.RWMutex
	v             map[K]*Container[T]
	expire        time.Duration
	sleepInterval time.Duration
}

//Get a value --
func (s *Cache[T, V]) Get(k V) (t T, err error) {
	s.m.RLock()
	defer s.m.RUnlock()
	if v, ok := s.v[k]; ok {
		return v.Load(), nil
	}
	return t, ErrKey
}

func (s *Cache[T, V]) Where(m func(T) bool) (v []T, err error) {
	return s.memQuery(m)
}

func (s *Cache[T, V]) memQuery(m func(T) bool) (v []T, err error) {
	v = make([]T, 0)
	s.m.RLock()
	defer s.m.RUnlock()
	for _, k := range s.GetKeys() {
		val := s.v[k].Load()
		if m(val) {
			v = append(v, val)
		}
	}
	return
}

//Set a value
func (s *Cache[T, V]) Set(k V, v T) error {
	s.m.RLock()
	if val, ok := s.v[k]; ok {
		val.Store(v)
		s.m.RUnlock()
		return nil
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	s.v[k] = s.newContainer(v)
	return nil
}

func (s *Cache[T, V]) Exists(k V) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

//Delete a value
func (s *Cache[T, V]) Delete(k V) (err error) {

	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

	return nil
}

func (s *Cache[T, V]) GetKeys(fromDb ...bool) (out []V) {
	s.m.RLock()
	defer s.m.RUnlock()
	out = make([]V, len(s.v))
	var i int
	for k := range s.v {
		out[i] = k
		i++
	}
	return out
}

func New[T any, V comparable](m ...map[V]T) (s *Cache[T, V]) {
	if len(m) > 0 {
		s = &Cache[T, V]{m: new(sync.RWMutex), v: make(map[V]*Container[T])}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &Cache[T, V]{m: new(sync.RWMutex), v: make(map[V]*Container[T])}
}

func (s *Cache[T, V]) janitor() {
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

func (s *Cache[T, V]) WithExpiration(e time.Duration) *Cache[T, V] {
	s.expire = e
	if e > (30 * time.Second) {
		s.sleepInterval = e
	} else {
		s.sleepInterval = 30 * time.Second
	}
	go s.janitor()
	return s
}

func (c *Cache[T, V]) newContainer(v T) *Container[T] {
	return &Container[T]{m: &sync.RWMutex{}, v: v, deadline: time.Now().Add(c.expire)}
}

func (s *Cache[T, V]) unCache(k V) (err error) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)
	return
}

func (s *Cache[T, V]) Filter(f func(T) bool) map[V]T {
	keys := s.GetKeys()
	out := make(map[V]T)
	s.m.RLock()
	defer s.m.RUnlock()
	for _, k := range keys {
		value := s.v[k]
		if f(value.Load()) {
			out[k] = value.Load()
		}
	}
	return out
}

func (s *Cache[T, V]) Values() []T {
	out := make([]T, 0, len(s.v))
	s.m.RLock()
	defer s.m.RUnlock()
	for _, k := range s.GetKeys() {
		out = append(out, s.v[k].Load())
	}
	return out
}

func (s *Cache[T, V]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *Cache[T, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

func (s *Cache[T, V]) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &s.v)
}

func (s *Cache[T, V]) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.v)
}
