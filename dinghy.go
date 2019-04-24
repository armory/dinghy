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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/armory-io/monitoring/log/formatters"
	"github.com/armory-io/monitoring/log/hooks"
	"github.com/armory/plank"

	"github.com/armory/dinghy/pkg/cache"
	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/settings"
	"github.com/armory/dinghy/pkg/util"
	"github.com/armory/dinghy/pkg/web"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func newRedisOptions(redisOptions settings.Redis) *redis.Options {
	url := strings.TrimPrefix(redisOptions.BaseURL, "redis://")
	return &redis.Options{
		Addr:     url,
		Password: redisOptions.Password,
		DB:       0,
	}
}

func main() {
	config, err := settings.LoadSettings()
	if err != nil {
		log.Fatalf("failed to load configuration: %s", err.Error())
	}

	if config.Logging.File != "" {
		f, err := os.OpenFile(config.Logging.File, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0764)
		if err != nil {
			log.Fatalf("Couldn't open log file")
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
		log.Fatalf("Invalid log level: " + logLevelStr)
	}
	log.SetLevel(logLevel)
	if config.Logging.Remote.Enabled {
		if err := setupRemoteLogging(log.StandardLogger(), config.Logging); err != nil {
			log.Warnf("unable to setup remote log forwarding: %s", err.Error())
		}
	}

	// New API client; nil arg uses default HTTP client.
	client := plank.NewAuthenticated(config.Fiat.AuthUser, nil)

	// Update the base URLs based on config
	client.URLs["orca"] = config.Orca.BaseURL
	client.URLs["front50"] = config.Front50.BaseURL

	// Create the EventClient
	ctx := context.Background()
	ec := events.NewEventClient(ctx, config)

	redis := cache.NewRedisCache(newRedisOptions(config.Redis))
	api := web.NewWebAPI(config, redis, client, ec)

	log.Info("Dinghy started.")
	log.Info(http.ListenAndServe(":8081", api.Router()))
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
