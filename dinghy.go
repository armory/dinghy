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
	dinghy_hcl "github.com/armory-io/dinghy/pkg/parsers/hcl"
	"os"
	"os/exec"

	// Open Core Dinghy
	dinghy_yaml "github.com/armory-io/dinghy/pkg/parsers/yaml"
	dinghy "github.com/armory/dinghy/cmd"

	"github.com/armory-io/dinghy/pkg/notifiers"
	"github.com/armory-io/dinghy/pkg/settings"
	settings_dinghy "github.com/armory/dinghy/pkg/settings"
	logr "github.com/sirupsen/logrus"
)

func main() {
	// Load default settings and execute liquibase script
	dinghySettings, err := settings_dinghy.LoadSettings()
	executeLiquibase(dinghySettings)

	log, api := dinghy.Setup()
	moreConfig, err := settings.LoadExtraSettings(api.Config)
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
	dinghy.Start(log, api, api.Config)
}


func executeLiquibase(settings *settings_dinghy.Settings) {
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
				fmt.Fprintf(os.Stdout,"cmd output: %v", string(cmd))
				fmt.Fprintf(os.Stdout,"Execution of /liquibase/liquibase-upgrade.sh failed: %v", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "Execution of /liquibase/liquibase-upgrade.sh succeeded: %v", string(cmd))

		} else {
			fmt.Fprintf(os.Stdout,"Something failed reading /liquibase/liquibase-upgrade.sh - %v", err)
			os.Exit(1)
		}
	}
}