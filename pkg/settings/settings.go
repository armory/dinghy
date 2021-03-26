package settings

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"context"
	"github.com/armory/dinghy/pkg/web"
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/armory/go-yaml-tools/pkg/spring"
)

type ExtSettings struct {
	// Notifiers configurations
	Notifiers notifierConfig `json:"notifiers,omitempty" yaml:"notifiers"`
	// Metrics configurations
	Metrics metricsConfig `json:"metrics,omitempty" yaml:"metrics"`
	// Dinghy OSS configuration
	Settings  *global.Settings
}

type notifierConfig struct {
	// Slack configuration
	// More info: https://docs.armory.io/docs/armory-admin/dinghy-enable/#slack-notifications
	Slack slackOpts `json:"slack,omitempty" yaml:"slack"`
	// Github comments validation, when a validation is done it will add a comment with the full log
	Github githubOpts `json:"github,omitempty" yaml:"github"`
}

type metricsConfig struct {
	NewRelic newRelicOpts `json:"newRelic,omitempty" yaml:"newRelic"`
}

type newRelicOpts struct{
	ApiKey string `json:"apiKey,omitempty" yaml:"apiKey"`
	ApplicationName string `json:"applicationName,omitempty" yaml:"applicationName"`
}
type slackOpts struct {
	// Enabled flag
	Enabled string `json:"enabled" yaml:"enabled"`
	// Slack channel configured for notifications
	Channel string `json:"channel" yaml:"channel"`
}

type githubOpts struct {
	// Enabled flag, by default is true
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

func LoadExtraSettings(s *global.Settings) (*ExtSettings, error) {
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

func (s ExtSettings) TraceExtract() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqWithTraceContext := r
			if s.Metrics.NewRelic.ApiKey != "" {
				tc := web.ExtractTraceContextHeaders(r.Header)
				if tc.TraceID != "" {
					cv := context.WithValue(r.Context(), web.TraceContextKey{}, tc)
					reqWithTraceContext = r.WithContext(cv)
				}
			}
			next.ServeHTTP(w, reqWithTraceContext)
		})
	}
}
