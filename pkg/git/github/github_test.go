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

package github

import (
	"errors"
	"testing"

	"github.com/armory/dinghy/pkg/git"
	"github.com/stretchr/testify/assert"
)

type GitHubTest struct {
	contents string
	endpoint string
	token    string
	err      error
}

func (g *GitHubTest) DownloadContents(org, repo, path, branch string) (string, error) {
	return g.contents, g.err
}

func (g *GitHubTest) CreateStatus(status *Status, org, repo, ref string) error {
	return g.err
}

func (g *GitHubTest) GetEndpoint() string {
	return g.endpoint
}

func (g *GitHubTest) GetToken() string {
	return g.token
}

func TestGitHubSlugger(t *testing.T) {

	cases := []struct {
		event            map[string]interface{}
		expectedErr      error
		expectedRepoInfo *git.RepositoryInfo
	}{
		{
			event: map[string]interface{}{
				"foo": "bar",
			},
			expectedErr: errors.New("unable to find html url from push event"),
		},
		{
			event: map[string]interface{}{
				"html_url": "https://github.com/armory/dinghy",
			},
			expectedRepoInfo: &git.RepositoryInfo{
				Org:  "armory",
				Repo: "dinghy",
				Type: "github",
			},
		},
	}

	cfg := Config{}
	for _, c := range cases {
		info, err := cfg.Slug(c.event)
		if err != nil {
			if c.expectedErr == nil {
				t.Fatalf("got an error but wasn't expecting one")
			}

			assert.Equal(t, err, c.expectedErr)
		}

		assert.Equal(t, c.expectedRepoInfo, info)
	}
}
