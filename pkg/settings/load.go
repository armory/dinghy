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
)

// S is the global settings structure
// In order to support legacy installs, S has only had additive changes made to it
// TODO: Remove S when customers are all using dinghy with halyard
var S = Settings{
	DinghyFilename:    "dinghyfile",
	TemplateRepo:      "dinghy-templates",
	AutoLockPipelines: "true",
	GitHubCredsPath:   util.GetenvOrDefault("GITHUB_TOKEN_PATH", os.Getenv("HOME")+"/.armory/cache/github-creds.txt"),
	GithubEndpoint:    "https://api.github.com",
	StashCredsPath:    util.GetenvOrDefault("STASH_TOKEN_PATH", os.Getenv("HOME")+"/.armory/cache/stash-creds.txt"),
	StashEndpoint:     "http://localhost:7990/rest/api/1.0",
	Logging: logging{
		File:  "",
		Level: "INFO",
	},
	spinnakerSupplied: spinnakerSupplied{
		Orca: spinnakerService{
			Enabled: "true",
			BaseURL: util.GetenvOrDefault("ORCA_BASE_URL", "http://orca:8083"),
		},
		Front50: spinnakerService{
			Enabled: "true",
			BaseURL: util.GetenvOrDefault("FRONT50_BASE_URL", "http://front50:8080"),
		},
		Fiat: fiat{
			spinnakerService: spinnakerService{
				Enabled: "false",
				BaseURL: util.GetenvOrDefault("FIAT_BASE_URL", "http://fiat:7003"),
			},
			AuthUser: "",
		},
		Redis: redis{
			Host:     util.GetenvOrDefault("REDIS_HOST", "redis"),
			Port:     util.GetenvOrDefault("REDIS_PORT", "6379"),
			Password: util.GetenvOrDefault("REDIS_PASSWORD", ""),
		},
	},
}

func init() {
	// Read in settings for dinghy and spinnaker profiles
	springConfig, err := loadProfiles()
	if err != nil {
		return
	}

	// Overwrite S's initial values with those stored in springConfig
	// TODO: Remove this code when customers are all using dinghy with halyard
	if err := mergo.Merge(&S, springConfig, mergo.WithOverride); err != nil {
		log.Errorf("failed to merge custom config with default: %s", err.Error())
		return
	}

	// If Github token not passed directly
	// Required for backwards compatibility
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
	// Required for backwards compatibility
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

	// Required for backwards compatibility
	if S.Deck.BaseURL == "" && S.SpinnakerUIURL != "" {
		log.Warn("Spinnaker UI URL should be set with ${services.deck.baseUrl}")
		S.Deck.BaseURL = S.SpinnakerUIURL
	}

	// Take the FiatUser setting if fiat is enabled (coming from hal settings)
	if S.Fiat.Enabled == "true" && S.FiatUser != "" {
		S.Fiat.AuthUser = S.FiatUser
	}

	c, _ := json.Marshal(S)
	log.Infof("The following settings have been loaded: %v", string(c))

}

func loadProfiles() (Settings, error) {
	// var s Settings
	var config Settings
	propNames := []string{"spinnaker", "dinghy"}
	c, err := spring.LoadDefault(propNames)
	if err != nil {
		log.Errorf("Could not load yaml configs - %v", err)
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
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		log.Errorf("Could not Unmarshall yaml configs into Settings - %v", err)
		return config, err
	}
	return config, nil
}
