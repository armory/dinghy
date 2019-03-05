package main

import (
	"fmt"
	"github.com/armory-io/monitoring/log/formatters"
	"github.com/armory-io/monitoring/log/hooks"
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

func newRedisOptions(redisOptions settings.Redis) *redis.Options {
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisOptions.Host, redisOptions.Port),
		Password: redisOptions.Password,
		DB:       0,
	}
}

func main() {
	config, err := settings.LoadSettings()
	if err != nil {
		log.Fatalf("failed to load configuration: %s", err.Error())
	}
	settings.ReplaceGlobals(*config)

	if config.Logging.File != "" {
		f, err := os.OpenFile(config.Logging.File, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0764)
		if err != nil {
			panic("Couldn't open log file")
		}
		log.SetOutput(f)
	} else {
		log.SetOutput(os.Stdout)
	}
	logLevelStr := util.GetenvOrDefault("DEBUG_LEVEL", "info")
	if config.Logging.Level != "" {
		logLevelStr = strings.ToLower(config.Logging.Level)
		log.Info("Debug level set to ", config.Logging.Level, " from settings")
	}
	logLevel, err := log.ParseLevel("DEBUG")
	if err != nil {
		log.Panic("Invalid log level : " + logLevelStr)
	}
	log.SetLevel(logLevel)
	if config.Logging.Remote.Enabled {
		if err := setupRemoteLogging(log.StandardLogger(), config.Logging); err != nil {
			log.Warnf("unable to setup remote log forwarding: %s", err.Error())
		}
	}

	log.Info("Dinghy started.")
	web.GlobalCache = cache.NewRedisCache(newRedisOptions(config.Redis))
	log.Fatal(http.ListenAndServe(":8081", web.Router()))
}

func setupRemoteLogging(l *log.Logger, loggingConfig settings.Logging) error {
	var hostname string
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = os.Getenv("HOSTNAME")
	}

	if hostname == "" {
		return fmt.Errorf("hostname could not be resolved.")
	}

	formatter, err := formatters.NewHttpLogFormatter(
		hostname,
		loggingConfig.Remote.CustomerID,
		loggingConfig.Remote.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to instantiate remote log forwarding: %s", err.Error())
	}

	if loggingConfig.Remote.Endpoint == "" {
		return fmt.Errorf("remote log forwarding is enabled but not loging.remote.endpoint is not set.")
	}

	l.AddHook(&hooks.HttpDebugHook{
		Endpoint:  loggingConfig.Remote.Endpoint,
		LogLevels: log.AllLevels,
		Formatter: formatter,
	})
	return nil
}
