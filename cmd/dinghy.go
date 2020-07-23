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

package dinghy

import (
	"context"
	"fmt"
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/armory/go-yaml-tools/pkg/tls/server"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/armory/dinghy/pkg/debug"

	"github.com/armory-io/monitoring/log/formatters"
	"github.com/armory-io/monitoring/log/hooks"
	"github.com/armory/plank/v3"

	"github.com/armory/dinghy/pkg/cache"
	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/settings"
	"github.com/armory/dinghy/pkg/util"
	"github.com/armory/dinghy/pkg/web"
	"github.com/go-redis/redis"
	logr "github.com/sirupsen/logrus"
)

func newRedisOptions(redisOptions settings.Redis) *redis.Options {
	url := strings.TrimPrefix(redisOptions.BaseURL, "redis://")
	return &redis.Options{
		MaxRetries: 5,
		Addr:       url,
		Password:   redisOptions.Password,
		DB:         0,
	}
}

func Setup() (*logr.Logger, *web.WebAPI) {
	log := logr.New()
	config, err := settings.LoadSettings()
	if err != nil {
		log.Fatalf("failed to load configuration: %s", err.Error())
	}

	if config.Logging.File != "" {
		f, err := os.OpenFile(config.Logging.File, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
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
	logLevel, err := logr.ParseLevel(logLevelStr)
	if err != nil {
		log.Fatalf("Invalid log level: " + logLevelStr)
	}
	log.SetLevel(logLevel)
	if config.Logging.Remote.Enabled {
		if err := setupRemoteLogging(logr.StandardLogger(), config.Logging); err != nil {
			log.Warnf("unable to setup remote log forwarding: %s", err.Error())
		}
	}

	client := setupPlankClient(config, log)
	clientReadOnly := util.PlankReadOnly{Plank: client}

	// Create the EventClient
	ctx, cancel := context.WithCancel(context.Background())
	ec := events.NewEventClient(ctx, config)

	// spawn stop thread
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-stop
		log.Info("Stopping Dinghy")
		cancel()
		os.Exit(1)
	}()

	redisClient := cache.NewRedisCache(newRedisOptions(config.Redis), log, ctx, stop)
	if _, err := redisClient.Client.Ping().Result(); err != nil {
		log.Fatalf("Redis Server at %s could not be contacted: %v", config.Redis.BaseURL, err)
	}
	redisClientReadOnly := cache.RedisCacheReadOnly{
		Client: redisClient.Client,
		Logger: redisClient.Logger,
	}
	api := web.NewWebAPI(config, redisClient, client, ec, log, &redisClientReadOnly, clientReadOnly)
	api.AddDinghyfileUnmarshaller(&dinghyfile.DinghyJsonUnmarshaller{})
	if config.ParserFormat == "json" {
		api.SetDinghyfileParser(dinghyfile.NewDinghyfileParser(&dinghyfile.PipelineBuilder{}))
	}
	return log, api
}

func setupPlankClient(settings *settings.Settings, log *logr.Logger) *plank.Client {
	var httpClient *http.Client
	if log.Level == logr.DebugLevel {
		httpClient = debug.NewInterceptorHttpClient(log, &settings.Http, true)
	} else {
		httpClient = settings.Http.NewClient()
	}
	client := plank.New(plank.WithClient(httpClient),
		plank.WithFiatUser(settings.Fiat.AuthUser))

	// Update the base URLs based on config
	client.URLs["orca"] = settings.Orca.BaseURL
	client.URLs["front50"] = settings.Front50.BaseURL
	return client
}

func AddUnmarshaller(u dinghyfile.DinghyJsonUnmarshaller, api *web.WebAPI) {
	api.AddDinghyfileUnmarshaller(u)
}

func Start(log *logr.Logger, api *web.WebAPI, settings2 *settings.Settings) {
	log.Infof("Dinghy starting on %s", settings2.Server.GetAddr())
	if err := server.NewServer(&settings2.Server).Start(api.Router()); err != nil {
		log.Fatal(err)
	}
}

func setupRemoteLogging(l *logr.Logger, loggingConfig settings.Logging) error {
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
		LogLevels: logr.AllLevels,
		Formatter: formatter,
	})
	return nil
}
