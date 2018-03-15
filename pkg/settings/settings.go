// Package settings is a single place to put all of the application settings.
package settings

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/armory-io/dinghy/pkg/util"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Settings contains all information needed to startup and run the dinghy service
type Settings struct {
	TemplateOrg       string `json:"templateOrg" yaml:"templateOrg"`
	DinghyFilename    string `json:"dinghyFilename" yaml:"dinghyFilename"`
	TemplateRepo      string `json:"templateRepo" yaml:"templateRepo"`
	AutoLockPipelines string `json:"autoLockPipelines" yaml:"autoLockPipelines"`
	SpinnakerAPIURL   string `json:"spinAPIUrl" yaml:"spinAPIUrl"`
	SpinnakerUIURL    string `json:"spinUIUrl" yaml:"spinUIUrl"`
	CertPath          string `json:"certPath" yaml:"certPath"`
	GitHubCredsPath   string `json:"githubCredsPath" yaml:"githubCredsPath"`
	GitHubToken       string
	StashCredsPath    string `json:"stashCredsPath" yaml:"stashCredsPath"`
	StashUsername     string
	StashToken        string
	StashEndpoint     string `json:"stashEndpoint" yaml:"stashEndpoint"`
}

// S is the global settings structure
var S = Settings{
	TemplateOrg:       "armory-io",
	DinghyFilename:    "dinghyfile",
	TemplateRepo:      "dinghy-templates",
	AutoLockPipelines: "true",
	SpinnakerAPIURL:   "https://spinnaker.armory.io:8085",
	SpinnakerUIURL:    "https://spinnaker.armory.io",
	CertPath:          util.GetenvOrDefault("CLIENT_CERT_PATH", os.Getenv("HOME")+"/.armory/cache/client.pem"),
	GitHubCredsPath:   util.GetenvOrDefault("GITHUB_TOKEN_PATH", os.Getenv("HOME")+"/.armory/cache/github-creds.txt"),
	StashCredsPath:    util.GetenvOrDefault("STASH_TOKEN_PATH", os.Getenv("HOME")+"/.armory/cache/stash-creds.txt"),
	StashEndpoint:     "http://localhost:7990/rest/api/1.0",
}

// If we got a DINGHY_CONFIG file as part of env, parse what's there into settings
// else initialize with default (Armory) values
func init() {
	var s Settings
	configFile := util.GetenvOrDefault("DINGHY_CONFIG", "/opt/spinnaker/dinghy-local.yml")
	if _, err := os.Stat(configFile); err == nil {
		bytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Errorf("Unable to open config file: %v", err)
			return
		}
		err = yaml.Unmarshal(bytes, &s)
		if err != nil {
			log.Errorf("Unable to parse config file: %v", err)
			return
		}
		log.Infof("Configured with settings from file: ", configFile)
		S = s
	} else {
		log.Info("Config file ", configFile, " not present falling back to default settings")
	}

	// load github api token
	if _, err := os.Stat(S.GitHubCredsPath); err == nil {
		creds, err := ioutil.ReadFile(S.GitHubCredsPath)
		if err != nil {
			panic(err)
		}
		c := strings.Split(strings.TrimSpace(string(creds)), ":")
		if len(c) < 2 {
			panic("github creds file should have format 'username:token'")
		}
		S.GitHubToken = c[1]
		log.Info("Successfully loaded github api creds")
	}

	// load stash api creds
	if _, err := os.Stat(S.StashCredsPath); err == nil {
		creds, err := ioutil.ReadFile(S.StashCredsPath)
		if err != nil {
			panic(err)
		}
		c := strings.Split(strings.TrimSpace(string(creds)), ":")
		if len(c) < 2 {
			panic("stash creds file should have format 'username:token'")
		}
		S.StashToken = c[0]
		S.StashToken = c[1]
		log.Info("Successfully loaded github api creds")
	}
}
