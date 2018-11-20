package cache

import (
	"testing"

	"fmt"

	"github.com/armory-io/dinghy/pkg/util"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func connectToRedis() *RedisCache {
	host := util.GetenvOrDefault("REDIS_HOST", "redis")
	port := util.GetenvOrDefault("REDIS_PORT", "6379")

	c := NewRedisCache(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: util.GetenvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	})

	c.Clear()
	return c
}

func TestRedisCache(t *testing.T) {
	c := connectToRedis()

	_, err := c.Ping().Result()
	if err != nil {
		t.Skip("Could not connect to Redis; skipping test")
	}

	c.SetDeps("df1", []string{"mod1", "mod2"})
	c.SetDeps("mod1", []string{"mod3", "mod4"})
	assert.EqualValuesf(t, []string{"df1"}, c.GetRoots("mod4"), "mod4 should have roots df1")

	c.SetDeps("mod1", []string{"mod3"})
	assert.EqualValuesf(t, []string{}, c.GetRoots("mod4"), "mod4 should have no roots")
}
