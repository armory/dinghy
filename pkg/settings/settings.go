// Package settings is a single place to put all of the application settings.
package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/armory-io/dinghy/pkg/util"
	"github.com/armory/go-yaml-tools/pkg/spring"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var YAMLConfig Settings
var S Settings

// Settings contains all information needed to startup and run the dinghy service
type Settings struct {
	TemplateOrg       string           `json:"templateOrg,omitempty" yaml:"templateOrg"`
	DinghyFilename    string           `json:"dinghyFilename,omitempty" yaml:"dinghyFilename"`
	TemplateRepo      string           `json:"templateRepo,omitempty" yaml:"templateRepo"`
	AutoLockPipelines string           `json:"autoLockPipelines,omitempty" yaml:"autoLockPipelines"`
	SpinnakerUIURL    string           `json:"spinUIUrl,omitempty" yaml:"spinUIUrl"`
	GitHubCredsPath   string           `json:"githubCredsPath,omitempty" yaml:"githubCredsPath"`
	GitHubToken       string           `json:"githubToken,omitempty" yaml:"githubToken"`
	GithubEndpoint    string           `json:"githubEndpoint,omitempty" yaml:"githubEndpoint"`
	StashCredsPath    string           `json:"stashCredsPath,omitempty" yaml:"stashCredsPath"`
	StashUsername     string           `json:"stashUsername,omitempty" yaml:"stashUsername"`
	StashToken        string           `json:"stashToken,omitempty" yaml:"stashToken"`
	StashEndpoint     string           `json:"stashEndpoint,omitempty" yaml:"stashEndpoint"`
	RedisServer       string           `json:"redisServer,omitempty" yaml:"redisServer"`
	RedisPassword     string           `json:"redisPassword,omitempty" yaml:"redisPassword"`
	Orca              spinnakerService `json:"orca,omitempty" yaml:"orca"`
	Front50           spinnakerService `json:"front50,omitempty" yaml:"front50"`
	Fiat              spinnakerService `json:"fiat,omitempty" yaml:"fiat"`
	Logging           logging          `json:"logging,omitempty" yaml:"logging"`
	FiatUser          string           `json:"fiatUser,omitempty" yaml:"fiatUser"`
}

type spinnakerService struct {
	Enabled  *bool  `json:"enabled" yaml:"enabled"`
	BaseURL  string `json:"baseUrl,omitempty" yaml:"baseUrl"`
	AuthUser string `json:"authUser,omitempty" yaml:"authUser"`
}

type logging struct {
	File  string `json:"file,omitempty" yaml:"file"`
	Level string `json:"level,omitempty" yaml:"level"`
}

func init() {
	springConfig, err := loadProfiles()
	if err != nil {
		return
	}
	s, _ := json.Marshal(springConfig)
	log.Infof("SpringConfig: %s", string(s))

}

// If we got a DINGHY_CONFIG file as part of env, parse what's there into settings
// else initialize with default (Armory) values
func init_old() {
	var s Settings
	configFile := util.GetenvOrDefault("DINGHY_CONFIG", "/opt/spinnaker/config/dinghy-local.yml")

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
		log.Infof("Configured with settings from file: %s", configFile)

		// mergo merges 2 like structs together
		if err := mergo.Merge(&S, s, mergo.WithOverride); err != nil {
			log.Errorf("failed to merge custom config with default: %s", err.Error())
			return
		}

	} else {
		log.Infof("Config file %s not present falling back to default settings", configFile)
	}

	// If Github token not passed directly
	if S.GitHubToken == "" {
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
	}

	// If Stash token not passed directly
	if S.StashToken == "" || S.StashUsername == "" {
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
			S.StashUsername = c[0]
			S.StashToken = c[1]
			log.Info("Successfully loaded stash api creds")
		}
	}

	// Take the FiatUser setting if fiat is enabled (coming from hal settings)
	if *(S.Fiat.Enabled) && S.FiatUser != "" {
		S.Fiat.AuthUser = S.FiatUser
	}
}

func loadProfiles() (Settings, error) {
	// var s Settings
	var config Settings
	propNames := []string{"spinnaker", "dinghy"}
	c, err := spring.LoadDefault(propNames)
	if err != nil {
		log.Errorf("Could not load yaml conifgs - %v", err)
		return config, err
	}
	// c is map[string]interface{} but we want it as Settings
	// so marshall to []byte as intermediate step
	bytes, err := json.Marshal(&c)
	if err != nil {
		log.Errorf("Could not marshall yaml configs - %v", err)
		return config, err
	}
	// and now unmarshall as Settings
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Errorf("Could not Unmarshall yaml configs into Settings - %v", err)
	}
	log.Infof("Using settings: %v", string(bytes))

	return config, nil
}

func inititalizeSettings() {
	// S is the legacy global settings structure
	t := new(bool)
	f := new(bool)

	*t = true
	*f = false

	S = Settings{
		TemplateOrg:       "armory-io",
		DinghyFilename:    "dinghyfile",
		TemplateRepo:      "dinghy-templates",
		AutoLockPipelines: "true",
		SpinnakerUIURL:    "https://spinnaker.armory.io",
		GitHubCredsPath:   util.GetenvOrDefault("GITHUB_TOKEN_PATH", os.Getenv("HOME")+"/.armory/cache/github-creds.txt"),
		GithubEndpoint:    "https://api.github.com",
		StashCredsPath:    util.GetenvOrDefault("STASH_TOKEN_PATH", os.Getenv("HOME")+"/.armory/cache/stash-creds.txt"),
		StashEndpoint:     "http://localhost:7990/rest/api/1.0",
		Logging: logging{
			File:  "",
			Level: "INFO",
		},
		Orca: spinnakerService{
			Enabled: t,
			BaseURL: util.GetenvOrDefault("ORCA_BASE_URL", "http://orca:8083"),
		},
		Front50: spinnakerService{
			Enabled: t,
			BaseURL: util.GetenvOrDefault("FRONT50_BASE_URL", "http://front50:8080"),
		},
		Fiat: spinnakerService{
			Enabled:  f,
			BaseURL:  util.GetenvOrDefault("FIAT_BASE_URL", "http://fiat:7003"),
			AuthUser: "",
		},
	}
}
