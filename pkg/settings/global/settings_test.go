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

package global

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettings_GetRepoConfig(t *testing.T) {
	cases := map[string]struct {
		settings Settings
		provider string
		repo     string
		expected *RepoConfig
	}{
		"happy path": {
			settings: Settings{
				RepoConfig: []RepoConfig{
					{
						Provider: "github",
						Repo:     "ghrepo",
						Branch:   "ghbranch",
					},
					{
						Provider: "bitbucket",
						Repo:     "bbrepo",
						Branch:   "bbbranch",
					},
				},
			},
			provider: "bitbucket",
			repo:     "bbrepo",
			expected: &RepoConfig{
				Provider: "bitbucket",
				Repo:     "bbrepo",
				Branch:   "bbbranch",
			},
		},
		"not found": {
			settings: Settings{
				RepoConfig: []RepoConfig{
					{
						Provider: "github",
						Repo:     "ghrepo",
						Branch:   "ghbranch",
					},
					{
						Provider: "bitbucket",
						Repo:     "bbrepo",
						Branch:   "bbbranch",
					},
				},
			},
			provider: "stash",
			repo:     "repo",
			expected: nil,
		},
		"no repo configuration": {
			settings: Settings{RepoConfig: []RepoConfig{}},
			expected: nil,
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			actual := c.settings.GetRepoConfig(c.provider, c.repo)
			assert.Equal(t, c.expected, actual)
		})
	}
}
