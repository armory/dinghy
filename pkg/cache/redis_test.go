package cache

import (
	"testing"

	"fmt"
	"github.com/armory-io/dinghy/pkg/util"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func TestBlahBlah(t *testing.T) {
	options := &redis.Options{}
	redisHost := util.GetenvOrDefault("REDIS_HOST", "redis")
	redisPort := util.GetenvOrDefault("REDIS_PORT", "6379")
	options.Addr = fmt.Sprintf("%s:%s", redisHost, redisPort)
	options.Password = util.GetenvOrDefault("REDIS_PASSWORD", "")
	options.DB = 0

	c := NewRedisCache(options)
	c.Clear()

	c.SetDeps("df1", []string{"mod1", "mod2"})
	c.SetDeps("mod1", []string{"mod3", "mod4"})
	assert.EqualValuesf(t, []string{"df1"}, c.GetRoots("mod4"), "mod4 should have roots df1")

	c.SetDeps("mod1", []string{"mod3"})
	assert.EqualValuesf(t, []string{}, c.GetRoots("mod4"), "mod4 should have no roots")
}
