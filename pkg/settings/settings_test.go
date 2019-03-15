package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigureSettings(t *testing.T) {
	defaultsWithOverridesExpected := NewDefaultSettings()
	defaultsWithOverridesExpected.Redis.Host = "12345"
	defaultsWithOverridesExpected.Redis.Port = "6789"

	cases := map[string]struct {
		defaults     Settings
		overrides    Settings
		expected     Settings
		expectsError bool
	}{
		"happy path": {
			defaults: Settings{
				GitHubToken: "12345",
			},
			overrides: Settings{
				GitHubToken: "45678",
			},
			expected: Settings{
				GitHubToken: "45678",
			},
		},
		"defaults no overrides": {
			defaults:  NewDefaultSettings(),
			overrides: Settings{},
			expected:  NewDefaultSettings(),
		},
		"defaults with spinnaker settings overridden": {
			defaults: NewDefaultSettings(),
			overrides: Settings{
				spinnakerSupplied: spinnakerSupplied{
					Redis: Redis{
						Host:     "12345",
						Port:     "6789",
						Password: "",
					},
				},
			},
			expected: defaultsWithOverridesExpected,
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			s, err := configureSettings(c.defaults, c.overrides)
			if c.expectsError {
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, c.expected, *s)
		})
	}
}

func TestDecodeProfilesToSettings(t *testing.T) {
	cases := map[string]struct {
		input        map[string]interface{}
		expected     Settings
		expectsError bool
	}{
		"happy path": {
			input: map[string]interface{}{
				"redis": map[string]interface{}{
					"host": "12345",
				},
			},
			expected: Settings{
				spinnakerSupplied: spinnakerSupplied{
					Redis: Redis{
						Host: "12345",
					},
				},
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			var decoded Settings
			err := decodeProfilesToSettings(c.input, &decoded)
			if c.expectsError {
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, c.expected, decoded)

		})
	}
}
