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
	"github.com/armory/dinghy/pkg/database"
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/armory/dinghy/pkg/execution"
	"github.com/armory/dinghy/pkg/logevents"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/armory/go-yaml-tools/pkg/tls/server"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/armory/dinghy/pkg/debug"

	"github.com/armory-io/monitoring/log/formatters"
	"github.com/armory-io/monitoring/log/hooks"
	"github.com/armory/plank/v4"

	"github.com/armory/dinghy/pkg/cache"
	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/util"
	"github.com/armory/dinghy/pkg/web"
	"github.com/go-redis/redis"
	logr "github.com/sirupsen/logrus"
)

func newRedisOptions(redisOptions global.Redis) *redis.Options {
	url := strings.TrimPrefix(redisOptions.BaseURL, "redis://")
	return &redis.Options{
		MaxRetries: 5,
		Addr:       url,
		Password:   redisOptions.Password,
		DB:         0,
	}
}

func Setup(sourceConfiguration source.SourceConfiguration, log *logr.Logger) (*logr.Logger, *web.WebAPI) {
	// We need to initialize the configuration for the start-up.
	config, err := sourceConfiguration.LoadSetupSettings(log)
	if err != nil {
		log.Fatal(fmt.Errorf("an error occurred when trying to load configurations: %w", err))
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
	ec := events.NewEventClient(ctx, config, sourceConfiguration.IsMultiTenant())

	// spawn stop thread
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-stop
		log.Info("Stopping Dinghy")
		cancel()
		os.Exit(1)
	}()

	var api *web.WebAPI
	var logEventsClient logevents.LogEventsClient
	var persitenceManager dinghyfile.DependencyManager
	var persitenceManagerReadOnly dinghyfile.DependencyManager

	// Full SQL mode
	if config.SQL.Enabled && !config.SQL.EventLogsOnly {

		sqlClient, sqlerr := database.NewMySQLClient(&database.SQLConfig{
			DbUrl:    config.SQL.BaseUrl,
			User:     config.SQL.User,
			Password: config.SQL.Password,
			DbName:   config.SQL.DatabaseName,
		}, log, ctx, stop)

		if sqlerr != nil {
			log.Fatalf("SQL Server at %s could not be contacted: %v", config.SQL.BaseUrl, err)
		}

		sqlClientReadOnly := database.SQLReadOnly{
			Client: sqlClient,
			Logger: sqlClient.Logger,
		}

		logEventsClient = &(logevents.LogEventSQLClient{SQLClient: sqlClient, MinutesTTL: config.LogEventTTLMinutes})
		persitenceManager = sqlClient
		persitenceManagerReadOnly = &sqlClientReadOnly

		redisClient := cache.NewRedisCache(newRedisOptions(config.SpinnakerSupplied.Redis), log, ctx, stop, false)

		var migration execution.Execution

		migration = &execution.RedisToSQLMigration{
			Settings:   config,
			Logger:     log,
			RedisCache: redisClient,
			SQLClient:  sqlClient,
		}

		migration.Execute()
		migration.Finalize()

	} else if config.SQL.Enabled && config.SQL.EventLogsOnly {
		// Hybrid SQL mode just for eventlogs
		sqlClient, sqlerr := database.NewMySQLClient(&database.SQLConfig{
			DbUrl:    config.SQL.BaseUrl,
			User:     config.SQL.User,
			Password: config.SQL.Password,
			DbName:   config.SQL.DatabaseName,
		}, log, ctx, stop)

		if sqlerr != nil {
			log.Fatalf("SQL Server at %s could not be contacted: %v", config.SQL.BaseUrl, err)
		}

		redisClient := cache.NewRedisCache(newRedisOptions(config.SpinnakerSupplied.Redis), log, ctx, stop, true)
		if _, err := redisClient.Client.Ping().Result(); err != nil {
			log.Fatalf("Redis Server at %s could not be contacted: %v", config.SpinnakerSupplied.Redis.BaseURL, err)
		}

		redisClientReadOnly := cache.RedisCacheReadOnly{
			Client: redisClient.Client,
			Logger: redisClient.Logger,
		}

		logEventsClient = &(logevents.LogEventSQLClient{SQLClient: sqlClient, MinutesTTL: config.LogEventTTLMinutes})
		persitenceManager = redisClient
		persitenceManagerReadOnly = &redisClientReadOnly

	} else {
		// Redis mode
		redisClient := cache.NewRedisCache(newRedisOptions(config.SpinnakerSupplied.Redis), log, ctx, stop, true)
		if _, err := redisClient.Client.Ping().Result(); err != nil {
			log.Fatalf("Redis Server at %s could not be contacted: %v", config.SpinnakerSupplied.Redis.BaseURL, err)
		}

		redisClientReadOnly := cache.RedisCacheReadOnly{
			Client: redisClient.Client,
			Logger: redisClient.Logger,
		}

		logEventsClient = logevents.LogEventRedisClient{RedisClient: redisClient, MinutesTTL: config.LogEventTTLMinutes}
		persitenceManager = redisClient
		persitenceManagerReadOnly = &redisClientReadOnly

	}

	api = web.NewWebAPI(sourceConfiguration, persitenceManager, ec, log, persitenceManagerReadOnly, &clientReadOnly, logEventsClient, log)
	api.MetricsHandler = new(web.NoOpMetricsHandler)
	api.AddDinghyfileUnmarshaller(&dinghyfile.DinghyJsonUnmarshaller{})
	if config.ParserFormat == "json" {
		api.SetDinghyfileParser(dinghyfile.NewDinghyfileParser(&dinghyfile.PipelineBuilder{}))
	}
	return log, api
}

func setupPlankClient(settings *global.Settings, log *logr.Logger) *plank.Client {
	var httpClient *http.Client
	if log.Level == logr.DebugLevel {
		httpClient = debug.NewInterceptorHttpClient(log, &settings.Http, true)
	} else {
		httpClient = settings.Http.NewClient()
	}
	client := plank.New(plank.WithClient(httpClient),
		plank.WithFiatUser(settings.SpinnakerSupplied.Fiat.AuthUser))

	// Update the base URLs based on config
	client.URLs["orca"] = settings.SpinnakerSupplied.Orca.BaseURL
	client.URLs["front50"] = settings.SpinnakerSupplied.Front50.BaseURL
	return client
}

func AddUnmarshaller(u dinghyfile.DinghyJsonUnmarshaller, api *web.WebAPI) {
	api.AddDinghyfileUnmarshaller(u)
}

func Start(log *logr.Logger, api *web.WebAPI, settings2 *global.Settings) {
	log.Infof("Dinghy starting on %s", settings2.Server.GetAddr())
	if err := server.NewServer(&settings2.Server).Start(api.MuxRouter); err != nil {
		log.Fatal(err)
	}
}

func setupRemoteLogging(l *logr.Logger, loggingConfig global.Logging) error {
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
