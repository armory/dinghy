package cache

import (
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/go-redis/redis"
)

// RedisCacheStore maintains a dependency graph inside Redis
type RedisCacheStore redis.Client

var redisOptions redis.Options

func init() {
	redisOptions.Addr = settings.S.RedisServer
	redisOptions.Password = settings.S.RedisPassword
	redisOptions.DB = 0
}

// NewRedisCacheStore initializes a new cache
func NewRedisCacheStore() RedisCacheStore {
	return RedisCacheStore{}
}

func (c RedisCacheStore) SetDeps(parent string, deps ...string) {
	// TODO
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
