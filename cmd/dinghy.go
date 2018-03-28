package main

import (
	"net/http"
	"os"

	"fmt"
	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/util"
	"github.com/armory-io/dinghy/pkg/web"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func makeRedisOptions() *redis.Options {
	options := &redis.Options{}

	redisHost := util.GetenvOrDefault("REDIS_HOST", "redis")
	redisPort := util.GetenvOrDefault("REDIS_PORT", "6379")
	options.Addr = fmt.Sprintf("%s:%s", redisHost, redisPort)
	options.Password = util.GetenvOrDefault("REDIS_PASSWORD", "")
	options.DB = 0

	return options
}

func main() {
	log.SetOutput(os.Stdout)
	logLevelStr := util.GetenvOrDefault("DEBUG_LEVEL", "info")
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		log.Panic("Invalid log level : " + logLevelStr)
	}
	log.SetLevel(logLevel)
	log.Info("Dinghy started.")

	cache.C = cache.NewRedisCache(makeRedisOptions())
	log.Fatal(http.ListenAndServe(":8081", web.Router()))
}
