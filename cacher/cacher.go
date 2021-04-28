/**
Implement a very simple cache for demo
*/
package cacher

import "sync"

type Cacher interface {
	Get(string) (string, bool)
	Set(string, string)
}

type cache struct {
	storage map[string]string
	lock    sync.RWMutex
}

// NewCacher create a cache instance
func NewCacher() Cacher {
	return &cache{
		storage: make(map[string]string),
	}
}

// Get get cache from key
func (c *cache) Get(key string) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v, ok := c.storage[key]
	return v, ok
}

// Set set cache key with value
func (c *cache) Set(key string, value string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.storage[key] = value
}
