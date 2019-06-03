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
	"github.com/sirupsen/logrus"
	"testing"

	"fmt"

	"github.com/armory/dinghy/pkg/util"
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
	}, logrus.New())

	c.Clear()
	return c
}

func TestRedisCache(t *testing.T) {
	c := connectToRedis()

	_, err := c.Client.Ping().Result()
	if err != nil {
		t.Skip("Could not connect to Redis; skipping test")
	}

	c.SetDeps("df1", []string{"mod1", "mod2"})
	c.SetDeps("mod1", []string{"mod3", "mod4"})
	assert.EqualValuesf(t, []string{"df1"}, c.GetRoots("mod4"), "mod4 should have roots df1")

	c.SetDeps("mod1", []string{"mod3"})
	assert.EqualValuesf(t, []string{}, c.GetRoots("mod4"), "mod4 should have no roots")
}
