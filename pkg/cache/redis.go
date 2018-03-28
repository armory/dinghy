package cache

import (
	"fmt"
	"strings"

	"github.com/armory-io/dinghy/pkg/util"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// RedisCacheStore maintains a dependency graph inside Redis
type RedisCacheStore redis.Client

var redisOptions redis.Options

func init() {
	redisHost := util.GetenvOrDefault("REDIS_HOST", "redis")
	redisPort := util.GetenvOrDefault("REDIS_PORT", "6379")
	redisOptions.Addr = fmt.Sprintf("%s:%s", redisHost, redisPort)
	redisOptions.Password = util.GetenvOrDefault("REDIS_PASSWORD", "")
	redisOptions.DB = 0
}

func compileKey(keys ...string) string {
	return fmt.Sprintf("Armory:%s", strings.Join(keys, ":"))
}

// NewRedisCacheStore initializes a new cache
func NewRedisCacheStore() *RedisCacheStore {
	return (*RedisCacheStore)(redis.NewClient(&redisOptions))
}

func (c RedisCacheStore) SetDeps(parent string, deps ...string) {
	key := compileKey("dinghy", "children", parent)

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

	key = compileKey("dinghy", "children", parent)
	c.SRem(key, depsToDelete...)
	c.SAdd(key, depsToAdd...)

	for _, dep := range depsToDelete {
		key = compileKey("dinghy", "parents", dep.(string))
		c.SRem(key, parent)
	}

	for _, dep := range depsToAdd {
		key = compileKey("dinghy", "parents", dep.(string))
		c.SAdd(key, parent)
	}
}

func (c RedisCacheStore) UpstreamURLs(url string) (upstreams, roots []string) {
	upstreams = make([]string, 0)
	roots = make([]string, 0)

	visited := map[string]bool{}
	q := make([]string, 0)
	q = append(q, url)

	for len(q) > 0 {
		curr := q[0]
		q = q[1:]

		visited[curr] = true

		key := compileKey("dinghy", "parents", curr)
		parents, err := c.SMembers(key).Result()
		if err != nil {
			log.Error(err)
			break
		}

		if curr != url {
			upstreams = append(upstreams, curr)
			if len(parents) == 0 {
				roots = append(roots, curr)
			}
		}

		for _, parent := range parents {
			if _, exists := visited[parent]; !exists {
				q = append(q, parent)
				visited[parent] = true
			}
		}
	}

	return
}

// Dump prints the cache, used for debugging
func (c RedisCacheStore) Dump() {
	// TODO
}

// Clear clears everything
func (c RedisCacheStore) Clear() {
	keys, _ := c.Keys("Armory:dinghy:children:*").Result()
	c.Del(keys...)

	keys, _ = c.Keys("Armory:dinghy:parents:*").Result()
	c.Del(keys...)
}
