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
		c.items[key] = c.queue.PushFront(value)
		if c.queue.Len() > c.capacity {
			li = c.queue.Back()
			key, ok = mapKey(c.items, li)
			if ok {
				delete(c.items, key)
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
	}
}

func mapKey(m map[Key]*ListItem, value *ListItem) (key Key, ok bool) {
	for k, v := range m {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}
