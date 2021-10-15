package gocache

import (
	"sync"
	"time"
)

type container struct {
	*sync.RWMutex
	v        Object
	deadline time.Time
}

func (c *container) load() Object {
	c.RLock()
	defer c.RUnlock()
	return c.v
}

func (c *container) store(v Object) {
	c.Lock()
	defer c.Unlock()
	c.v = v
}

func (s *ItemCache) newContainer(v Object) *container {
	return &container{
		RWMutex:  new(sync.RWMutex),
		v:        v,
		deadline: time.Now().Add(s.expire),
	}
}
