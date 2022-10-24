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
func (c *Cache[K, V]) Get(k K) (t V, err error) {
	c.m.RLock()
	defer c.m.RUnlock()
	if v, ok := c.v[k]; ok {
		return v.Load(), nil
	}
	return t, ErrKey
}

// Set a value
func (c *Cache[K, V]) Set(k K, v V) error {
	c.m.RLock()
	if val, ok := c.v[k]; ok {
		val.Store(v)
		c.m.RUnlock()
		return nil
	}
	c.m.RUnlock()
	c.m.Lock()
	defer c.m.Unlock()
	c.v[k] = c.newContainer(v)
	return nil
}

func (c *Cache[K, V]) Exists(k K) bool {
	c.m.RLock()
	defer c.m.RUnlock()
	_, ok := c.v[k]
	return ok
}

// Delete a value
func (c *Cache[K, V]) Delete(k K) (err error) {

	c.m.Lock()
	defer c.m.Unlock()
	delete(c.v, k)

	return nil
}

func (c *Cache[K, V]) GetKeys() (out []K) {
	c.m.RLock()
	defer c.m.RUnlock()
	out = make([]K, len(c.v))
	var i int
	for k := range c.v {
		out[i] = k
		i++
	}
	return out
}

func New[K comparable, V any](m ...map[K]V) (c *Cache[K, V]) {
	if len(m) > 0 {
		c = &Cache[K, V]{m: new(sync.RWMutex), v: make(map[K]*Container[V])}
		for k, v := range m[0] {
			c.Set(k, v)
		}
		return c
	}
	return &Cache[K, V]{m: new(sync.RWMutex), v: make(map[K]*Container[V])}
}

func (c *Cache[K, V]) janitor() {
	for {
		time.Sleep(c.sleepInterval)
		now := time.Now()
		for _, k := range c.GetKeys() {
			c.m.RLock()
			if v, ok := c.v[k]; ok {
				if v.deadline.Before(now) {
					c.m.RUnlock()
					c.unCache(k)
				} else {
					c.m.RUnlock()
				}
			} else {
				c.m.RUnlock()
			}
		}
	}
}

func (c *Cache[K, V]) WithExpiration(e time.Duration) *Cache[K, V] {
	c.expire = e
	if e > (30 * time.Second) {
		c.sleepInterval = e
	} else {
		c.sleepInterval = 30 * time.Second
	}
	go c.janitor()
	return c
}

func (c *Cache[K, V]) newContainer(v V) *Container[V] {
	return &Container[V]{m: &sync.RWMutex{}, v: v, deadline: time.Now().Add(c.expire)}
}

func (c *Cache[K, V]) unCache(k K) (err error) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.v, k)
	return
}

func (c *Cache[K, V]) Filter(f func(V) bool) map[K]V {
	keys := c.GetKeys()
	out := make(map[K]V)
	c.m.RLock()
	defer c.m.RUnlock()
	for _, k := range keys {
		value := c.v[k]
		if f(value.Load()) {
			out[k] = value.Load()
		}
	}
	return out
}

func (c *Cache[K, V]) Values() []V {
	out := make([]V, 0, len(c.v))
	c.m.RLock()
	defer c.m.RUnlock()
	for _, k := range c.GetKeys() {
		out = append(out, c.v[k].Load())
	}
	return out
}

func (c *Cache[K, V]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &c.v)
}

func (c *Cache[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.v)
}

func (c *Cache[K, V]) UnmarshalYAML(b []byte) error {
	return yaml.Unmarshal(b, &c.v)
}

func (c *Cache[K, V]) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(c.v)
}
