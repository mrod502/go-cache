package gocache

import (
	"sync"
	"time"
)

type container struct {
	*sync.RWMutex
	v        interface{}
	deadline time.Time
}

func (c *container) load() interface{} {
	c.RLock()
	defer c.RUnlock()
	return c.v
}

func (c *container) store(v interface{}) {
	c.Lock()
	defer c.Unlock()
	c.v = v
}

func (s *InterfaceCache) newContainer(v interface{}) *container {
	return &container{
		RWMutex:  new(sync.RWMutex),
		v:        v,
		deadline: time.Now().Add(s.expire),
	}
}
