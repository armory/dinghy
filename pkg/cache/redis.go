package cache

import (
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/go-redis/redis"
	"fmt"
	"strings"
	log "github.com/sirupsen/logrus"
)

// RedisCacheStore maintains a dependency graph inside Redis
type RedisCacheStore redis.Client

var redisOptions redis.Options

func init() {
	redisOptions.Addr = settings.S.RedisServer
	redisOptions.Password = settings.S.RedisPassword
	redisOptions.DB = 0
}

func compileKey(keys ...string) string {
	return fmt.Sprintf("Armory:%s", strings.Join(keys, ":"))
}

// NewRedisCacheStore initializes a new cache
func NewRedisCacheStore() RedisCacheStore {
	return RedisCacheStore{}
}

func (c RedisCacheStore) SetDeps(parent string, deps ...string) {
	key := compileKey("dinghy", "children", parent)

	currentDeps, err := c.SMembers(key).Result()
	if err != nil {
		log.Error(err)
		return
	}

	toDelete := make(map[string]bool, 0)
	for _, currentDep := range currentDeps {
		toDelete[currentDep] = true
	}

	for _, dep := range deps {
		delete(toDelete, dep)
	}

	keysToDelete := make([]interface{}, 0)
	for key := range toDelete {
		keysToDelete = append(keysToDelete, key)
	}

	keysToAdd := make([]interface{}, 0)
	for _, key := range deps {
		keysToAdd = append(keysToAdd, key)
	}

	key = compileKey("dinghy", "children", parent)
	c.SRem(key, keysToDelete...)
	c.SAdd(key, keysToAdd...)

	// TODO: finish this
}

// UpstreamNodes returns the upstream nodes and root nodes for a given node
func (c RedisCacheStore) UpstreamURLs(n *Node) (upstreams, roots []*Node) {
	// TODO
	return nil, nil
}

// Dump prints the cache, used for debugging
func (c RedisCacheStore) Dump() {
	// TODO
}
