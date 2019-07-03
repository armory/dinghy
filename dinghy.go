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
	dinghy_hcl "github.com/armory-io/dinghy/pkg/parsers/hcl"
	// Open Core Dinghy
	dinghy_yaml "github.com/armory-io/dinghy/pkg/parsers/yaml"
	dinghy "github.com/armory/dinghy/cmd"

	"github.com/armory-io/dinghy/pkg/notifiers"
	"github.com/armory-io/dinghy/pkg/settings"
)

func main() {
	log, api := dinghy.Setup()
	moreConfig, err := settings.LoadExtraSettings(api.Config)
	if err != nil {
		log.Errorf("Error loading additional settings: %s", err.Error())
	}
	if moreConfig.Notifiers.Slack.IsEnabled() {
		log.Infof("Slack notifications enabled, sending to %s", moreConfig.Notifiers.Slack.Channel)
	} else {
		log.Info("Slack notifications disabled/not configured")
	}
	if moreConfig.Notifiers.Slack.IsEnabled() {
		api.AddNotifier(notifiers.NewSlackNotifier(moreConfig))
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
	dinghy.Start(log, api)
}
