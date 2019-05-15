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

package cache

import (
	"fmt"
	"strings"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// RedisCache maintains a dependency graph inside Redis
type RedisCache redis.Client

func compileKey(keys ...string) string {
	return fmt.Sprintf("Armory:dinghy:%s", strings.Join(keys, ":"))
}

// NewRedisCache initializes a new cache
func NewRedisCache(redisOptions *redis.Options) *RedisCache {
	return (*RedisCache)(redis.NewClient(redisOptions))
}

// SetDeps sets dependencies for a parent
func (c *RedisCache) SetDeps(parent string, deps []string) {
	key := compileKey("children", parent)

	currentDeps, err := c.SMembers(key).Result()
	if err != nil {
		log.Error(err)
		return
	}

	// suppose current deps are (a, b, c) and new deps are (c, d, e)

	// generate (a, b)
	toDelete := make(map[string]bool, 0)
	for _, currentDep := range currentDeps {
		toDelete[currentDep] = true
	}
	for _, dep := range deps {
		delete(toDelete, dep)
	}
	depsToDelete := make([]interface{}, 0)
	for key := range toDelete {
		depsToDelete = append(depsToDelete, key)
	}

	// generate (d, e)
	toAdd := make(map[string]bool, 0)
	for _, dep := range deps {
		toAdd[dep] = true
	}
	for _, currentDep := range currentDeps {
		delete(toAdd, currentDep)
	}
	depsToAdd := make([]interface{}, 0)
	for _, key := range deps {
		depsToAdd = append(depsToAdd, key)
	}

	key = compileKey("children", parent)
	c.SRem(key, depsToDelete...)
	c.SAdd(key, depsToAdd...)

	for _, dep := range depsToDelete {
		key = compileKey("parents", dep.(string))
		c.SRem(key, parent)
	}

	for _, dep := range depsToAdd {
		key = compileKey("parents", dep.(string))
		c.SAdd(key, parent)
	}
}

// GetRoots grabs roots
func (c *RedisCache) GetRoots(url string) []string {
	roots := make([]string, 0)
	visited := map[string]bool{}

	for q := []string{url}; len(q) > 0; {
		curr := q[0]
		q = q[1:]

		visited[curr] = true

		key := compileKey("parents", curr)
		parents, err := c.SMembers(key).Result()
		if err != nil {
			log.Error(err)
			break
		}

		if curr != url && len(parents) == 0 {
			roots = append(roots, curr)
		}

		for _, parent := range parents {
			if _, exists := visited[parent]; !exists {
				q = append(q, parent)
				visited[parent] = true
			}
		}
	}

	return roots
}

// Clear clears everything
func (c *RedisCache) Clear() {
	keys, _ := c.Keys(compileKey("children", "*")).Result()
	c.Del(keys...)

	keys, _ = c.Keys(compileKey("parents", "*")).Result()
	c.Del(keys...)
}
