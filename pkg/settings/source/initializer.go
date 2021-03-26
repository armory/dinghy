package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/secret"
	"github.com/armory/go-yaml-tools/pkg/spring"
	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
)

type Initialize struct{}

func NewInitialize() *Initialize {
	return &Initialize{}
}

func (i *Initialize) Autoconfigure() (*global.Settings, error) {

	config := global.NewDefaultSettings()
	springConfig, err := i.loadProfiles()
	if err != nil {
		return nil, err
	}
	if err := mergo.Merge(&config, springConfig, mergo.WithOverride); err != nil {
		return nil, err
	}

	http := config.Http
	if err = http.Init(); err != nil {
		return nil, err
	}

	secretHandler, err := secret.NewSecretHandler(config.Secrets)
	if err != nil {
		return nil, err
	}

	if err := secretHandler.Decrypt(context.TODO(), &config); err != nil {
		return nil, fmt.Errorf("failed to decrypt secrets: %s", err)
	}

	return i.configureSettings(config)
}

func (i *Initialize) configureSettings(settings global.Settings) (*global.Settings, error) {

	// If Github token not passed directly
	// Required for backwards compatibility
	if settings.GitHubToken == "" {
		// load github api token
		if _, err := os.Stat(settings.GitHubCredsPath); err == nil {
			creds, err := ioutil.ReadFile(settings.GitHubCredsPath)
			if err != nil {
				return nil, err
			}
			c := strings.Split(strings.TrimSpace(string(creds)), ":")
			if len(c) < 2 {
				return nil, errors.New("github creds file should have format 'username:token'")
			}
			settings.GitHubToken = c[1]
			log.Info("Successfully loaded github api creds")
		}
	}

	// If Stash token not passed directly
	// Required for backwards compatibility
	if settings.StashToken == "" || settings.StashUsername == "" {
		// load stash api creds
		if _, err := os.Stat(settings.StashCredsPath); err == nil {
			creds, err := ioutil.ReadFile(settings.StashCredsPath)
			if err != nil {
				return nil, err
			}
			c := strings.Split(strings.TrimSpace(string(creds)), ":")
			if len(c) < 2 {
				return nil, errors.New("stash creds file should have format 'username:token'")
			}
			settings.StashUsername = c[0]
			settings.StashToken = c[1]
			log.Info("Successfully loaded stash api creds")
		}
	}

	// Required for backwards compatibility
	if settings.SpinnakerSupplied.Deck.BaseURL == "" && settings.SpinnakerUIURL != "" {
		log.Warn("Spinnaker UI URL should be set with ${services.deck.baseUrl}")
		settings.SpinnakerSupplied.Deck.BaseURL = settings.SpinnakerUIURL
	}

	// Take the FiatUser setting if fiat is enabled (coming from hal settings)
	if settings.SpinnakerSupplied.Fiat.Enabled == "true" && settings.FiatUser != "" {
		settings.SpinnakerSupplied.Fiat.AuthUser = settings.FiatUser
	}

	if settings.ParserFormat == "" {
		settings.ParserFormat = "json"
	}

	c, _ := json.Marshal(settings.Redacted())
	log.Infof("The following settings have been loaded: %v", string(c))

	return &settings, nil
}

func (i *Initialize) decodeProfilesToSettings(profiles map[string]interface{}, s *global.Settings) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           s,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(profiles)
}

func (i *Initialize) loadProfiles() (global.Settings, error) {
	// var s Settings
	var config global.Settings
	propNames := []string{"spinnaker", "dinghy"}
	c, err := spring.LoadDefault(propNames)
	if err != nil {
		return config, err
	}

	if err := i.decodeProfilesToSettings(c, &config); err != nil {
		return config, err
	}

	return config, nil
}
