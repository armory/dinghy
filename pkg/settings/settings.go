package settings

import (
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/armory/dinghy/pkg/settings"
	"github.com/armory/go-yaml-tools/pkg/spring"
)

type ExtSettings struct {
	Notifiers notifierConfig `json:"notifiers,omitempty" yaml:"notifiers"`
	Settings  *settings.Settings
}

type notifierConfig struct {
	Slack  slackOpts  `json:"slack,omitempty" yaml:"slack"`
	Github githubOpts `json:"github,omitempty" yaml:"github"`
}

type slackOpts struct {
	Enabled string `json:"enabled" yaml:"enabled"`
	Channel string `json:"channel" yaml:"channel"`
}

type githubOpts struct {
	Enabled string `json:"enabled" yaml:"enabled"`
}

func (s slackOpts) IsEnabled() bool {
	return (s.Channel != "") && (strings.ToLower(s.Enabled) == "true")
}

func (g githubOpts) IsEnabled() bool {
	if g.Enabled == "" {
		return true
	}
	return strings.ToLower(g.Enabled) == "true"
}

func LoadExtraSettings(s *settings.Settings) (*ExtSettings, error) {
	var cfg ExtSettings
	raw, err := spring.LoadDefault([]string{"spinnaker", "dinghy"})
	if err != nil {
		return &cfg, err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &cfg,
	})
	if err != nil {
		return &cfg, err
	}
	err = decoder.Decode(raw)
	cfg.Settings = s
	return &cfg, err
}
