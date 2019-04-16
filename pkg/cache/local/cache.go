/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

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
