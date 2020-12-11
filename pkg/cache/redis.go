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
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// RedisCache maintains a dependency graph inside Redis
type RedisCache struct {
	Client *redis.Client
	Logger *log.Entry
	ctx    context.Context
	stop   chan os.Signal
}

func CompileKey(keys ...string) string {
	return fmt.Sprintf("Armory:dinghy:%s", strings.Join(keys, ":"))
}

// NewRedisCache initializes a new cache
func NewRedisCache(redisOptions *redis.Options, logger *log.Logger, ctx context.Context, stop chan os.Signal, startMonitor bool) *RedisCache {
	rc := &RedisCache{
		Client: redis.NewClient(redisOptions),
		Logger: logger.WithFields(log.Fields{"cache": "redis"}),
		ctx:    ctx,
		stop:   stop,
	}
	if startMonitor {
		go rc.monitorWorker()
	}
	return rc
}

func (c *RedisCache) monitorWorker() {
	timer := time.NewTicker(10 * time.Second)
	count := 0
	for {
		select {
		case <-timer.C:
			if _, err := c.Client.Ping().Result(); err != nil {
				count++
				c.Logger.Errorf("Redis monitor failed %d times (5 max)", count)
				if count >= 5 {
					c.Logger.Error("Stopping dinghy because communication with redis failed")
					timer.Stop()
					c.stop <- syscall.SIGINT
				}
				continue
			}
			count = 0
		case <-c.ctx.Done():
			return
		}
	}
}

// SetDeps sets dependencies for a parent
func (c *RedisCache) SetDeps(parent string, deps []string) {
	loge := log.WithFields(log.Fields{"func": "SetDeps"})
	key := CompileKey("children", parent)

	currentDeps, err := c.Client.SMembers(key).Result()
	if err != nil {
		loge.WithFields(log.Fields{"operation": "get members", "key": key}).Error(err)
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

	//TODO:  if these redis operations fail, what happens?
	key = CompileKey("children", parent)
	if _, err := c.Client.SRem(key, depsToDelete...).Result(); err != nil {
		loge.WithFields(log.Fields{"operation": "child delete deps"}).Debug(err)
	}
	if _, err := c.Client.SAdd(key, depsToAdd...).Result(); err != nil {
		loge.WithFields(log.Fields{"operation": "child add deps"}).Debug(err)
	}

	for _, dep := range depsToDelete {
		key = CompileKey("parents", dep.(string))
		if _, err := c.Client.SRem(key, parent).Result(); err != nil {
			loge.WithFields(log.Fields{"operation": "delete deps"}).Debug(err)
		}
	}

	for _, dep := range depsToAdd {
		key = CompileKey("parents", dep.(string))
		if _, err := c.Client.SAdd(key, parent).Result(); err != nil {
			loge.WithFields(log.Fields{"operation": "add deps"}).Debug(err)
		}
	}
}

// GetRoots grabs roots
func (c *RedisCache) GetRoots(url string) []string {
	return returnRoots(c.Client, url)
}

func returnRoots(c *redis.Client, url string) []string {
	roots := make([]string, 0)
	visited := map[string]bool{}
	loge := log.WithFields(log.Fields{"func": "GetRoots"})

	for q := []string{url}; len(q) > 0; {
		curr := q[0]
		q = q[1:]

		visited[curr] = true

		key := CompileKey("parents", curr)
		parents, err := c.SMembers(key).Result()
		if err != nil {
			loge.WithFields(log.Fields{"operation": "parents", "key": key}).Error(err)
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

// Set RawData
func (c *RedisCache) SetRawData(url string, rawData string) error {
	loge := log.WithFields(log.Fields{"func": "SetRawData"})
	key := CompileKey("rawdata", url)

	status := c.Client.Set(key, rawData, 0)
	if status.Err() != nil {
		loge.WithFields(log.Fields{"operation": "set value", "key": key}).Error(status.Err())
		return status.Err()
	}
	return nil
}

func (c *RedisCache) GetRawData(url string) (string, error) {
	return returnRawData(c.Client, url)
}

func returnRawData(c *redis.Client, url string) (string, error) {
	loge := log.WithFields(log.Fields{"func": "GetRawData"})
	key := CompileKey("rawdata", url)

	stringCmd := c.Get(key)
	if stringCmd.Err() != nil {
		loge.WithFields(log.Fields{"operation": "get value", "key": key}).Error(stringCmd.Err())
		return "", stringCmd.Err()
	}
	return stringCmd.Result()
}

// Clear clears everything
func (c *RedisCache) Clear() {
	keys, _ := c.Client.Keys(CompileKey("children", "*")).Result()
	c.Client.Del(keys...)

	keys, _ = c.Client.Keys(CompileKey("parents", "*")).Result()
	c.Client.Del(keys...)
}


// Get all Dinghyfiles
func (c *RedisCache) GetAllDinghyfiles() []string {
	loge := log.WithFields(log.Fields{"func": "GetAllDinghyfiles"})
	key := CompileKey("parents", "*")
	var cursor uint64
	result := []string{}
	childrens := map[string]bool{}
	for {
		keys, nextcursor, err := c.Client.Scan(cursor, key, 1000).Result()
		cursor = nextcursor
		if err != nil {
			loge.WithFields(log.Fields{"operation": "scan key", "key": CompileKey("parents","*")}).Error(err)
			return result
		}
		for _, key := range keys {
			childrens[key] = true
		}

		if cursor == 0 {
			break
		}
	}

	for currentChildren, _ := range childrens {
		parents, errorNoKey := c.Client.SMembers(currentChildren).Result()
		if errorNoKey != nil {
			continue
		}
		for _, currentParent := range parents {
			compiledChildren := CompileKey("parents", currentParent)
			if _, ok := childrens[compiledChildren]; !ok {
				result = append(result, currentParent)
			}
		}
	}

	return result
}

// Get all childrens from url
func (c *RedisCache) GetChildren(url string) []string {
	loge := log.WithFields(log.Fields{"func": "GetChildren"})
	key := CompileKey("children", url)

	childrens, errorNoKey := c.Client.SMembers(key).Result()
	if errorNoKey != nil {
		loge.WithFields(log.Fields{"operation": "SMembers key", "key": key}).Error(errorNoKey)
		return []string{}
	}

	return childrens
}

