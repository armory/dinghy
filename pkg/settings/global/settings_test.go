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
		branch   string
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
						Provider: "github",
						Repo:     "ghrepo",
						Branch:   "ghanotherbranch",
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
			branch:   "bbbranch",
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
			branch:   "branch",
			expected: nil,
		},
		"no repo configuration": {
			settings: Settings{RepoConfig: []RepoConfig{}},
			expected: nil,
		},
		"when feature flag enabled, search by provider and repo and branch": {
			settings: Settings{
				MultipleBranchesEnabled: true,
				RepoConfig: []RepoConfig{
					{
						Provider: "github",
						Repo:     "ghrepo",
						Branch:   "main",
					},
					{
						Provider: "github",
						Repo:     "ghrepo",
						Branch:   "release",
					},
				},
			},
			provider: "github",
			repo:     "ghrepo",
			branch:   "featurebranch",
			expected: nil,
		},
		"when feature flag disabled, search just by provider and repo": {
			settings: Settings{
				MultipleBranchesEnabled: false,
				RepoConfig: []RepoConfig{
					{
						Provider: "github",
						Repo:     "ghrepo",
						Branch:   "main",
					},
					{
						Provider: "github",
						Repo:     "ghrepo",
						Branch:   "release",
					},
				},
			},
			provider: "github",
			repo:     "ghrepo",
			branch:   "featurebranch",
			expected: &RepoConfig{
				Provider: "github",
				Repo:     "ghrepo",
				Branch:   "main",
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			actual := c.settings.GetRepoConfig(c.provider, c.repo, c.branch)
			assert.Equal(t, c.expected, actual)
		})
	}
}
