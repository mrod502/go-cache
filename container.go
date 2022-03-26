package gocache

import (
	"sync"
	"time"
)

type icontainer struct {
	*sync.RWMutex
	v        interface{}
	deadline time.Time
}

func (c *icontainer) load() interface{} {
	c.RLock()
	defer c.RUnlock()
	return c.v
}

func (c *icontainer) store(v interface{}) {
	c.Lock()
	defer c.Unlock()
	c.v = v
}

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

func (s *ObjectCache) newContainer(v Object) *container {
	return &container{
		RWMutex:  new(sync.RWMutex),
		v:        v,
		deadline: time.Now().Add(s.expire),
	}
}

func (s *InterfaceCache) newiContainer(v interface{}) *icontainer {
	return &icontainer{
		RWMutex:  new(sync.RWMutex),
		v:        v,
		deadline: time.Now().Add(s.expire),
	}
}
