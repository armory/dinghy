package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/util"
	"github.com/armory-io/dinghy/pkg/web"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func newRedisOptions() *redis.Options {
	r := settings.S.Redis
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", r.Host, r.Port),
		Password: r.Password,
		DB:       0,
	}
}

func main() {
	if settings.S.Logging.File != "" {
		f, err := os.OpenFile(settings.S.Logging.File, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0764)
		if err != nil {
			panic("Couldn't open log file")
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}
	logLevelStr := util.GetenvOrDefault("DEBUG_LEVEL", "info")
	if settings.S.Logging.Level != "" {
		logLevelStr = strings.ToLower(settings.S.Logging.Level)
		log.Info("Debug level set to ", settings.S.Logging.Level, " from settings")
	}
	logLevel, err := log.ParseLevel("DEBUG")
	if err != nil {
		log.Panic("Invalid log level : " + logLevelStr)
	}
	log.SetLevel(logLevel)
	log.Info("Dinghy started.")

	web.GlobalCache = cache.NewRedisCache(newRedisOptions())
	log.Fatal(http.ListenAndServe(":8081", web.Router()))
}
