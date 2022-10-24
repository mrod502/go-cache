package gocache

import (
	"encoding/json"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Container[V any] struct {
	m        *sync.RWMutex
	v        V
	deadline time.Time
}

func NewContainer[V any](v V, deadline time.Time) *Container[V] {
	return &Container[V]{
		m:        &sync.RWMutex{},
		v:        v,
		deadline: deadline,
	}
}

func (c *Container[V]) Load() V {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.v
}
func (c *Container[V]) Store(v V) {
	c.m.Lock()
	defer c.m.Unlock()
	c.v = v
}

func (c *Container[V]) Expired() bool { return time.Now().UnixNano() > c.deadline.UnixNano() }

type Stringer interface {
	String() string
}

type Cache[K comparable, V any] struct {
	m             *sync.RWMutex
	v             map[K]*Container[V]
	expire        time.Duration
	sleepInterval time.Duration
}

// Get a value --
func (s *Cache[K, V]) Get(k K) (t V, err error) {
	s.m.RLock()
	defer s.m.RUnlock()
	if v, ok := s.v[k]; ok {
		return v.Load(), nil
	}
	return t, ErrKey
}

// Set a value
func (s *Cache[K, V]) Set(k K, v V) error {
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

func (s *Cache[K, V]) Exists(k K) bool {
	s.m.RLock()
	defer s.m.RUnlock()
	_, ok := s.v[k]
	return ok
}

// Delete a value
func (s *Cache[K, V]) Delete(k K) (err error) {

	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)

	return nil
}

func (s *Cache[K, V]) GetKeys() (out []K) {
	s.m.RLock()
	defer s.m.RUnlock()
	out = make([]K, len(s.v))
	var i int
	for k := range s.v {
		out[i] = k
		i++
	}
	return out
}

func New[K comparable, V any](m ...map[K]V) (s *Cache[K, V]) {
	if len(m) > 0 {
		s = &Cache[K, V]{m: new(sync.RWMutex), v: make(map[K]*Container[V])}
		for k, v := range m[0] {
			s.Set(k, v)
		}
		return s
	}
	return &Cache[K, V]{m: new(sync.RWMutex), v: make(map[K]*Container[V])}
}

func (s *Cache[K, V]) janitor() {
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

func (s *Cache[K, V]) WithExpiration(e time.Duration) *Cache[K, V] {
	s.expire = e
	if e > (30 * time.Second) {
		s.sleepInterval = e
	} else {
		s.sleepInterval = 30 * time.Second
	}
	go s.janitor()
	return s
}

func (c *Cache[K, V]) newContainer(v V) *Container[V] {
	return &Container[V]{m: &sync.RWMutex{}, v: v, deadline: time.Now().Add(c.expire)}
}

func (s *Cache[K, V]) unCache(k K) (err error) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.v, k)
	return
}

func (s *Cache[K, V]) Filter(f func(V) bool) map[K]V {
	keys := s.GetKeys()
	out := make(map[K]V)
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

func (s *Cache[K, V]) Values() []V {
	out := make([]V, 0, len(s.v))
	s.m.RLock()
	defer s.m.RUnlock()
	for _, k := range s.GetKeys() {
		out = append(out, s.v[k].Load())
	}
	return out
}

func (s *Cache[K, V]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &s.v)
}

func (s *Cache[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.v)
}

func (s *Cache[K, V]) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &s.v)
}

func (s *Cache[K, V]) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.v)
}
