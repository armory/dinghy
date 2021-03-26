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
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	dinghy_hcl "github.com/armory-io/dinghy/pkg/parsers/hcl"

	"github.com/armory-io/dinghy/pkg/notifiers"
	// Open Core Dinghy
	dinghy_yaml "github.com/armory-io/dinghy/pkg/parsers/yaml"
	"github.com/armory-io/dinghy/pkg/settings"
	dinghy "github.com/armory/dinghy/cmd"
	settings_dinghy "github.com/armory/dinghy/pkg/settings"
	global_settings_dinghy "github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/web"
	"github.com/newrelic/go-agent/v3/newrelic"
	logr "github.com/sirupsen/logrus"
)

type NewRelicMetricsHandler struct {
	app *newrelic.Application
}

func (nrm NewRelicMetricsHandler) WrapHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	return newrelic.WrapHandleFunc(nrm.app, pattern, handler)
}

func main() {
	// Load default settings and execute liquibase script
	dinghySettingsConfig, err := settings_dinghy.LoadSettings()
	if err != nil {
		log.Fatalf("could not load dinghy settings: %s", err.Error())
	}
	dinghySettings, err := dinghySettingsConfig.LoadSetupSettings()
	if err != nil {
		log.Fatalf("could not load local dinghy settings: %s", err.Error())
	}
	executeLiquibase(dinghySettings)

	log, api := dinghy.Setup()
	dinghySettings, err = api.SourceConfig.LoadSetupSettings()
	moreConfig, err := settings.LoadExtraSettings(dinghySettings)
	if err != nil {
		log.Errorf("Error loading additional settings: %s", err.Error())
	}
	if moreConfig.Notifiers.Slack.IsEnabled() {
		log.Infof("Slack notifications enabled, sending to %s", moreConfig.Notifiers.Slack.Channel)
		api.AddNotifier(notifiers.NewSlackNotifier(moreConfig))
	} else {
		log.Info("Slack notifications disabled/not configured")
	}

	if moreConfig.Notifiers.Github.IsEnabled() {
		log.Infof("Github notifications enabled")
		api.AddNotifier(notifiers.NewGithubNotifier(moreConfig))
	} else {
		log.Info("Github notifications disabled")
	}
	var app *newrelic.Application
	if moreConfig.Metrics.NewRelic.ApiKey != "" {
		if moreConfig.Metrics.NewRelic.ApplicationName == "" {
			log.Error("An NewRelic application name must be pecified")
		}
		app, err = newrelic.NewApplication(
			// Name your application
			newrelic.ConfigAppName(moreConfig.Metrics.NewRelic.ApplicationName),
			// Fill in your New Relic license key
			newrelic.ConfigLicense(moreConfig.Metrics.NewRelic.ApiKey),
			// Add logging:
			newrelic.ConfigDebugLogger(os.Stdout),
			// Optional: add additional changes to your configuration via a config function:
			func(cfg *newrelic.Config) {
				cfg.CustomInsightsEvents.Enabled = false
			},
		)
		// If an application could not be created then err will reveal why.
		if err != nil {
			fmt.Println("unable to create New Relic Application", err)
		}
		var mh NewRelicMetricsHandler
		mh.app = app
		api.MetricsHandler = mh
	} else {
		api.MetricsHandler = new(web.NoOpMetricsHandler)
	}
	switch moreConfig.Settings.ParserFormat {
	case "yaml":
		log.Info("Setting Dinghyfile parser to YAML")
		api.AddDinghyfileUnmarshaller(&dinghy_yaml.DinghyYaml{})
		api.SetDinghyfileParser(&dinghy_yaml.DinghyfileYamlParser{})
	case "hcl":
		log.Info("Settting Dinghyfile parser to HCL")
		api.AddDinghyfileUnmarshaller(&dinghy_hcl.DinghyHcl{})
		api.SetDinghyfileParser(&dinghy_hcl.DinghyfileHclParser{})
	}

	api.Client.EnableArmoryEndpoints()
	dinghy.Start(log, api, dinghySettings)
}

func executeLiquibase(settings *global_settings_dinghy.Settings) {
	log := logr.New()
	if settings.SQL.Enabled {
		log.Info("SQL.Enabled is true so /liquibase/liquibase-upgrade.sh will run")
		if _, err := os.Stat("/liquibase/liquibase-upgrade.sh"); err == nil {

			log.Info("Validating SQL configuration")
			if settings.SQL.BaseUrl == "" {
				log.Fatal("SQL.BaseUrl cannot be empty")
			}
			if settings.SQL.DatabaseName == "" {
				log.Fatal("SQL.DatabaseName cannot be empty")
			}
			if settings.SQL.User == "" {
				log.Fatal("SQL.User cannot be empty")
			}
			if settings.SQL.Password == "" {
				log.Fatal("SQL.Password cannot be empty")
			}

			cmd, err := exec.Command("/liquibase/liquibase-upgrade.sh", settings.SQL.BaseUrl,
				settings.SQL.DatabaseName, settings.SQL.User, settings.SQL.Password).CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stdout, "cmd output: %v", string(cmd))
				fmt.Fprintf(os.Stdout, "Execution of /liquibase/liquibase-upgrade.sh failed: %v", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "Execution of /liquibase/liquibase-upgrade.sh succeeded: %v", string(cmd))

		} else {
			fmt.Fprintf(os.Stdout, "Something failed reading /liquibase/liquibase-upgrade.sh - %v", err)
			os.Exit(1)
		}
	}
}
