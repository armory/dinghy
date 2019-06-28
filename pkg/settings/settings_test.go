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

package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigureSettings(t *testing.T) {
	defaultsWithOverridesExpected := NewDefaultSettings()
	defaultsWithOverridesExpected.Redis.BaseURL = "12345:6789"

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
				ParserFormat: "json",
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
						BaseURL:  "12345:6789",
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
					"baseUrl": "12345",
				},
			},
			expected: Settings{
				spinnakerSupplied: spinnakerSupplied{
					Redis: Redis{
						BaseURL: "12345",
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
