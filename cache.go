package cache

import (
	"container/list"
	"sync"
	"time"
)

type hitType int

const (
	found hitType = iota
	miss
	expired
)

type item struct {
	value interface{}
	exp   time.Time
	link  *list.Element
}

type Cache struct {
	sync.RWMutex
	Capacity   int
	Expiration time.Duration
	data       map[string]*item
	queue      *list.List
}

func (c *Cache) Size() int {
	return len(c.data)
}

func (c *Cache) Get(key string) (v interface{}, ok bool) {
	value, hit := c.get(key)
	switch hit {
	case found:
		v = value.value
		ok = true
	case miss:
	case expired:
		c.Remove(key)
	}
	return
}

func (c *Cache) get(key string) (*item, hitType) {
	c.RLock()
	defer c.RUnlock()
	value, ok := c.data[key]
	if !ok {
		return nil, miss
	}
	if c.Expiration != 0 || value.exp.Before(time.Now()) {
		return nil, expired
	}
	return value, found
}

func (c *Cache) Remove(key string) {
	c.Lock()
	defer c.Unlock()
	value, found := c.data[key]
	if found {
		c.queue.Remove(value.link)
		delete(c.data, key)
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if c.Capacity <= 0 {
		return
	}
	if c.data == nil {
		c.data = make(map[string]*item)
	}
	v, found := c.data[key]
	if found {
		v.value = value
		if c.Expiration > 0 {
			v.exp = time.Now().Add(c.Expiration)
		}
		return
	}
	for len(c.data) >= c.Capacity {
		front := c.queue.Front()
		key := front.Value.(string)
		delete(c.data, key)
		c.queue.Remove(front)
	}
	v = &item{
		value: value,
	}
	if c.Expiration > 0 {
		v.exp = time.Now().Add(c.Expiration)
	}
	c.data[key] = v
	c.queue.PushBack(key)
	v.link = c.queue.Back()
}
