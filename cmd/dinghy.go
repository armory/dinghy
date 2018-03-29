package main

import (
	"net/http"
	"os"

	"fmt"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/util"
	"github.com/armory-io/dinghy/pkg/web"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func makeRedisOptions() *redis.Options {
	host := util.GetenvOrDefault("REDIS_HOST", "redis")
	port := util.GetenvOrDefault("REDIS_PORT", "6379")

	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: util.GetenvOrDefault("REDIS_PASSWORD", ""),
		DB:       0,
	}
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

	dinghyfile.C = cache.NewRedisCache(makeRedisOptions())
	log.Fatal(http.ListenAndServe(":8081", web.Router()))
}
