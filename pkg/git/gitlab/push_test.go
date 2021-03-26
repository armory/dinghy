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

package gitlab

import (
	"encoding/json"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"testing"
	"time"
)

type commitsStruct []*struct {
	ID        string     `json:"id"`
	Message   string     `json:"message"`
	Timestamp *time.Time `json:"timestamp"`
	URL       string     `json:"url"`
	Author    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
	Removed  []string `json:"removed"`
}

type projectStruct struct {
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	AvatarURL         string                 `json:"avatar_url"`
	GitSSHURL         string                 `json:"git_ssh_url"`
	GitHTTPURL        string                 `json:"git_http_url"`
	Namespace         string                 `json:"namespace"`
	PathWithNamespace string                 `json:"path_with_namespace"`
	DefaultBranch     string                 `json:"default_branch"`
	Homepage          string                 `json:"homepage"`
	URL               string                 `json:"url"`
	SSHURL            string                 `json:"ssh_url"`
	HTTPURL           string                 `json:"http_url"`
	WebURL            string                 `json:"web_url"`
	Visibility        gitlab.VisibilityValue `json:"visibility"`
}

func loadExample(t *testing.T) []byte {
	content, err := ioutil.ReadFile("example_payload.json")
	require.Equal(t, err, nil)
	return content
}

func TestInSlice(t *testing.T) {
	testCases := map[string]struct {
		slice     []string
		valToFind string
		expected  bool
	}{
		"success": {
			slice:     []string{"app", "src", "dinghyfile"},
			valToFind: "dinghyfile",
			expected:  true,
		},
		"failure": {
			slice:     []string{"app", "src", "meh"},
			valToFind: "dinghyfile",
			expected:  false,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual := inSlice(tc.slice, tc.valToFind)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestContainsFile(t *testing.T) {
	testCases := map[string]struct {
		file     string
		push     *Push
		expected bool
	}{
		"added file": {
			file: "added-some-file",
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: commitsStruct{
						{
							Added:    []string{"added-some-file"},
							Modified: []string{},
						},
					},
				},
			},
			expected: true,
		},
		"modified file": {
			file: "modified-some-file",
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: commitsStruct{
						{
							Added:    []string{},
							Modified: []string{"modified-some-file"},
						},
					},
				},
			},
			expected: true,
		},
		"no commits": {
			file: "dinghyfile",
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: nil,
				},
			},
			expected: false,
		},
		"not in commits": {
			file: "some-other-file",
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: commitsStruct{
						{
							Added:    []string{"added-some-file"},
							Modified: []string{"modified-some-file"},
						},
					},
				},
			},
			expected: false,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual := tc.push.ContainsFile(tc.file)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestFiles(t *testing.T) {
	testCases := map[string]struct {
		push     *Push
		expected []string
	}{
		"files added": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: commitsStruct{
						{
							Added:    []string{"added-some-file", "added-another-file"},
							Modified: []string{},
						},
					},
				},
			},
			expected: []string{"added-some-file", "added-another-file"},
		},
		"files modified": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: commitsStruct{
						{
							Added:    []string{},
							Modified: []string{"modified-some-file", "modified-another-file"},
						},
					},
				},
			},
			expected: []string{"modified-some-file", "modified-another-file"},
		},
		"files added & modified": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: commitsStruct{
						{
							Added:    []string{"added-some-file", "added-another-file"},
							Modified: []string{"modified-some-file", "modified-another-file"},
						},
					},
				},
			},
			expected: []string{"added-some-file", "added-another-file", "modified-some-file", "modified-another-file"},
		},
		"none": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: commitsStruct{
						{
							Added:    []string{},
							Modified: []string{},
						},
					},
				},
			},
			expected: []string{},
		},
		"Null commits": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Commits: nil,
				},
			},
			expected: []string{},
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual := tc.push.Files()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestRepo(t *testing.T) {
	testCases := map[string]struct {
		push     *Push
		expected string
	}{
		"success": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Project: projectStruct{
						Name: "my-repo",
					},
				},
			},
			expected: "my-repo",
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual := tc.push.Repo()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestOrg(t *testing.T) {
	testCases := map[string]struct {
		push     *Push
		expected string
	}{
		"success": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Project: projectStruct{
						PathWithNamespace: "org/stuf/meh",
					},
				},
			},
			expected: "org",
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual := tc.push.Org()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestBranch(t *testing.T) {
	testCases := map[string]struct {
		push     *Push
		expected string
	}{
		"success": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "my-branch",
				},
			},
			expected: "my-branch",
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual := tc.push.Branch()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIsBranch(t *testing.T) {
	testCases := map[string]struct {
		push        *Push
		branchToTry string
		expected    bool
	}{
		"match: short name compare to long name": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "my-branch",
				},
			},
			branchToTry: "refs/heads/my-branch",
			expected:    true,
		},
		"match: long name compare to short name": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "refs/heads/my-branch",
				},
			},
			branchToTry: "my-branch",
			expected:    true,
		},
		"match: short same": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "my-branch",
				},
			},
			branchToTry: "my-branch",
			expected:    true,
		},
		"match: long same": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "refs/heads/my-branch",
				},
			},
			branchToTry: "refs/heads/my-branch",
			expected:    true,
		},
		"no match": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "refs/heads/my-branch",
				},
			},
			branchToTry: "some-other-branch",
			expected:    false,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual := tc.push.IsBranch(tc.branchToTry)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIsMaster(t *testing.T) {
	testCases := map[string]struct {
		push     *Push
		expected bool
	}{
		"true": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "refs/heads/master",
				},
			},
			expected: true,
		},
		"false": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "refs/heads/not-master",
				},
			},
			expected: false,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual := tc.push.IsMaster()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestParseWebhook(t *testing.T) {
	testCases := map[string]struct {
		push        *Push
		settings    *global.Settings
		body        []byte
		expectedErr bool
	}{
		"success": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "refs/heads/master",
				},
			},
			settings: &global.Settings{
				GitLabToken:    "token",
				GitLabEndpoint: "https://my-endpoint",
			},
			body:        loadExample(t),
			expectedErr: false,
		},
		"failed": {
			push: &Push{
				Event: &gitlab.PushEvent{
					Ref: "refs/heads/master",
				},
			},
			settings: &global.Settings{
				GitLabToken:    "token",
				GitLabEndpoint: "  https://my-endpoint",
			},
			body:        loadExample(t),
			expectedErr: true,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			_, err := tc.push.ParseWebhook(tc.settings, tc.body)

			var expectedEvent gitlab.PushEvent
			// make sure no error is returned
			require.Nil(t, json.Unmarshal(tc.body, &expectedEvent))

			if (err != nil) != tc.expectedErr {
				t.Errorf("ParseWebhook() error = %v, wantErr %v", err, tc.expectedErr)
				return
			}

			if !tc.expectedErr {
				assert.Equal(t, &expectedEvent, tc.push.Event)
			}
		})
	}
}
