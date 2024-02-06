package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
	Len() int
}

type lruCache struct {
	sync.RWMutex
	capacity int
	queue    List
	items    map[Key]*ListItem
	keys     map[*ListItem]Key
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	if c == nil {
		panic("cache is not initialized")
	}
	c.Lock()
	defer c.Unlock()
	_, ok := c.items[key]
	var li *ListItem
	keyFound := false
	if ok {
		li = c.items[key]
		li.Value = value
		c.queue.MoveToFront(li)
		keyFound = true
	} else {
		li = c.queue.PushFront(value)
		c.items[key] = li
		c.keys[li] = key
		if c.queue.Len() > c.capacity {
			li = c.queue.Back()
			key, ok = c.keys[li]
			if ok {
				delete(c.items, key)
				delete(c.keys, li)
			}
			c.queue.Remove(li)
		}
	}
	return keyFound
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	if c == nil {
		panic("cache is not initialized")
	}

	c.RLock()
	defer c.RUnlock()
	_, ok := c.items[key]
	var li *ListItem
	var value any
	keyFound := false
	if ok {
		li = c.items[key]
		value = li.Value
		c.queue.MoveToFront(li)
		keyFound = true
	}

	return value, keyFound
}

func (c *lruCache) Clear() {
	if c == nil {
		panic("cache is not initialized")
	}

	c.Lock()
	defer c.Unlock()

	for k, v := range c.items {
		c.queue.Remove(v)
		delete(c.items, k)
	}

	for k := range c.keys {
		delete(c.keys, k)
	}

	// new builtin function for maps since Go 1.21
	// clear(c.items)
}

func (c *lruCache) Len() int {
	return c.queue.Len()
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
		keys:     make(map[*ListItem]Key, capacity),
	}
}
