package source

import (
	"github.com/armory/dinghy/pkg/settings/lighthouse"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigureSettings(t *testing.T) {
	cases := map[string]struct {
		defaults     lighthouse.Settings
		overrides    lighthouse.Settings
		expected     func(*testing.T, *lighthouse.Settings)
		expectsError bool
	}{
		"happy path": {
			defaults: lighthouse.Settings{
				GitHubToken: "12345",
			},
			overrides: lighthouse.Settings{
				GitHubToken: "45678",
			},
			expected: func(t *testing.T, settings *lighthouse.Settings) {
				assert.Equal(t, settings, &lighthouse.Settings{
					ParserFormat: "json",
					GitHubToken:  "45678",
				})
			},
		},
		"defaults no overrides": {
			defaults:  lighthouse.NewDefaultSettings(),
			overrides: lighthouse.Settings{},
			expected: func(t *testing.T, settings *lighthouse.Settings) {
				assert.NotEmpty(t, settings)
			},
		},
		"defaults with spinnaker settings overridden": {
			defaults: lighthouse.NewDefaultSettings(),
			overrides: lighthouse.Settings{
				SpinnakerSupplied: lighthouse.SpinnakerSupplied{
					Redis: lighthouse.Redis{
						BaseURL:  "12345:6789",
						Password: "",
					},
				},
			},
			expected: func(t *testing.T, settings *lighthouse.Settings) {
				assert.Equal(t, "12345:6789", settings.SpinnakerSupplied.Redis.BaseURL)
				assert.Empty(t, settings.SpinnakerSupplied.Redis.Password)
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			i := &Initialize{}
			s, err := i.configureSettings(c.overrides)
			if !assert.Equal(t, c.expectsError, err != nil) {
				return
			}
			c.expected(t, s)
		})
	}
}

func TestDecodeProfilesToSettings(t *testing.T) {
	cases := map[string]struct {
		input        map[string]interface{}
		expected     lighthouse.Settings
		expectsError bool
	}{
		"happy path": {
			input: map[string]interface{}{
				"redis": map[string]interface{}{
					"baseUrl": "12345",
				},
			},
			expected: lighthouse.Settings{
				SpinnakerSupplied: lighthouse.SpinnakerSupplied{
					Redis: lighthouse.Redis{
						BaseURL: "12345",
					},
				},
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			i := &Initialize{}
			var decoded lighthouse.Settings
			err := i.decodeProfilesToSettings(c.input, &decoded)
			if c.expectsError {
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, c.expected, decoded)

		})
	}
}
