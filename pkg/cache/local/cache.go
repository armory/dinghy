// Package local provides a goroutine safe map for caching.
package local

import (
	"sync"
)

// Cache is a goroutine safe map.
type Cache struct {
	l sync.RWMutex
	m map[string]string
}

// Add a value to the cache.
func (c *Cache) Add(key, value string) {
	if c.m == nil {
		c.m = make(map[string]string)
	}
	c.l.Lock()
	c.m[key] = value
	c.l.Unlock()
}

// Get a value from the cache.
func (c *Cache) Get(key string) string {
	if c.m == nil {
		c.m = make(map[string]string)
		return ""
	}
	c.l.RLock()
	v := c.m[key]
	c.l.RUnlock()
	return v
}

// Len provides the size of the cache.
func (c *Cache) Len() int {
	return len(c.m)
}
