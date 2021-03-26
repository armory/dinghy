package source

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigureSettings(t *testing.T) {
	cases := map[string]struct {
		defaults     global.Settings
		overrides    global.Settings
		expected     func(*testing.T, *global.Settings)
		expectsError bool
	}{
		"happy path": {
			defaults: global.Settings{
				GitHubToken: "12345",
			},
			overrides: global.Settings{
				GitHubToken: "45678",
			},
			expected: func(t *testing.T, settings *global.Settings) {
				assert.Equal(t, settings, &global.Settings{
					ParserFormat: "json",
					GitHubToken:  "45678",
				})
			},
		},
		"defaults no overrides": {
			defaults:  global.NewDefaultSettings(),
			overrides: global.Settings{},
			expected: func(t *testing.T, settings *global.Settings) {
				assert.NotEmpty(t, settings)
			},
		},
		"defaults with spinnaker settings overridden": {
			defaults: global.NewDefaultSettings(),
			overrides: global.Settings{
				SpinnakerSupplied: global.SpinnakerSupplied{
					Redis: global.Redis{
						BaseURL:  "12345:6789",
						Password: "",
					},
				},
			},
			expected: func(t *testing.T, settings *global.Settings) {
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
		expected     global.Settings
		expectsError bool
	}{
		"happy path": {
			input: map[string]interface{}{
				"redis": map[string]interface{}{
					"baseUrl": "12345",
				},
			},
			expected: global.Settings{
				SpinnakerSupplied: global.SpinnakerSupplied{
					Redis: global.Redis{
						BaseURL: "12345",
					},
				},
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			i := &Initialize{}
			var decoded global.Settings
			err := i.decodeProfilesToSettings(c.input, &decoded)
			if c.expectsError {
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, c.expected, decoded)

		})
	}
}

func TestInitialize_Autoconfigure(t *testing.T) {
	util.CopyToLocalSpinnaker("testdata/dinghy-local.yml", "dinghy-local.yml")

	i := &Initialize{}
	config, err := i.Autoconfigure()

	assert.Nil(t, err)
	assert.Equal(t, "sample-org", config.TemplateOrg)
}
