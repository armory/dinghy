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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/armory/dinghy/pkg/git"
	"github.com/armory/dinghy/pkg/mock"
)

func TestSetCommitStatusSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Error(gomock.Any()).Times(0)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "")
	}))
	defer ts.Close()
	// TODO: Do not use global variable. This will lead to side-effects.
	p := Push{
		Config: Config{
			Endpoint: ts.URL,
		},
		Repository: Repository{Organization: "armory", Name: "dinghy"},
		Commits: []Commit{
			{
				ID: "ABC",
			},
		},
		Logger: logger,
	}

	// This shouldn't throw exceptions/panics
	p.SetCommitStatus(git.StatusPending, git.DefaultPendingMessage)
}

func TestSetCommitStatusFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Error(gomock.Any()).Times(1)

	// TODO: Do not use global variable. This will lead to side-effects.
	p := Push{
		Config: Config{
			Endpoint: "invalid-url",
		},
		Repository: Repository{Organization: "armory", Name: "dinghy"},
		Commits: []Commit{
			{
				ID: "ABC",
			},
		},
		Logger: logger,
	}

	// TODO: this doesn't actually test anything
	// This shouldn't throw exceptions/panics
	p.SetCommitStatus(git.StatusPending, git.DefaultPendingMessage)
}


func TestGetCommitStatus(t *testing.T) {

	tests := []struct {
		name string
		payload string
		err error
		status git.Status
		description string
		warning int
	}{
		{
			"payload_have_status",
			`[
    {
        "url": "https://api.github.com/repos/armory-io/dinghy-test-jossue/statuses/613be8796790c0e88bd9e69054894b762ef16c88",
        "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
        "id": 10532473715,
        "node_id": "MDEzOlN0YXR1c0NvbnRleHQxMDUzMjQ3MzcxNQ==",
        "state": "failure",
        "description": null,
        "target_url": null,
        "context": "PR Title Checker (becomes a merge/squash commit)",
        "created_at": "2020-08-19T21:17:02Z",
        "updated_at": "2020-08-19T21:17:02Z",
        "creator": {
            "login": "armory-jenkins",
            "id": 24237724,
            "node_id": "MDQ6VXNlcjI0MjM3NzI0",
            "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
            "gravatar_id": "",
            "url": "https://api.github.com/users/armory-jenkins",
            "html_url": "https://github.com/armory-jenkins",
            "followers_url": "https://api.github.com/users/armory-jenkins/followers",
            "following_url": "https://api.github.com/users/armory-jenkins/following{/other_user}",
            "gists_url": "https://api.github.com/users/armory-jenkins/gists{/gist_id}",
            "starred_url": "https://api.github.com/users/armory-jenkins/starred{/owner}{/repo}",
            "subscriptions_url": "https://api.github.com/users/armory-jenkins/subscriptions",
            "organizations_url": "https://api.github.com/users/armory-jenkins/orgs",
            "repos_url": "https://api.github.com/users/armory-jenkins/repos",
            "events_url": "https://api.github.com/users/armory-jenkins/events{/privacy}",
            "received_events_url": "https://api.github.com/users/armory-jenkins/received_events",
            "type": "User",
            "site_admin": false
        }
    },
    {
        "url": "https://api.github.com/repos/armory-io/dinghy-test-jossue/statuses/613be8796790c0e88bd9e69054894b762ef16c88",
        "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
        "id": 10532473053,
        "node_id": "MDEzOlN0YXR1c0NvbnRleHQxMDUzMjQ3MzA1Mw==",
        "state": "pending",
        "description": null,
        "target_url": null,
        "context": "PR Title Checker (becomes a merge/squash commit)",
        "created_at": "2020-08-19T21:16:58Z",
        "updated_at": "2020-08-19T21:16:58Z",
        "creator": {
            "login": "armory-jenkins",
            "id": 24237724,
            "node_id": "MDQ6VXNlcjI0MjM3NzI0",
            "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
            "gravatar_id": "",
            "url": "https://api.github.com/users/armory-jenkins",
            "html_url": "https://github.com/armory-jenkins",
            "followers_url": "https://api.github.com/users/armory-jenkins/followers",
            "following_url": "https://api.github.com/users/armory-jenkins/following{/other_user}",
            "gists_url": "https://api.github.com/users/armory-jenkins/gists{/gist_id}",
            "starred_url": "https://api.github.com/users/armory-jenkins/starred{/owner}{/repo}",
            "subscriptions_url": "https://api.github.com/users/armory-jenkins/subscriptions",
            "organizations_url": "https://api.github.com/users/armory-jenkins/orgs",
            "repos_url": "https://api.github.com/users/armory-jenkins/repos",
            "events_url": "https://api.github.com/users/armory-jenkins/events{/privacy}",
            "received_events_url": "https://api.github.com/users/armory-jenkins/received_events",
            "type": "User",
            "site_admin": false
        }
    },
    {
        "url": "https://api.github.com/repos/armory-io/dinghy-test-jossue/statuses/613be8796790c0e88bd9e69054894b762ef16c88",
        "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
        "id": 10532473015,
        "node_id": "MDEzOlN0YXR1c0NvbnRleHQxMDUzMjQ3MzAxNQ==",
        "state": "success",
        "description": "No changes in dinghyfile.",
        "target_url": "",
        "context": "dinghy",
        "created_at": "2020-08-19T21:16:57Z",
        "updated_at": "2020-08-19T21:16:57Z",
        "creator": {
            "login": "armory-jenkins",
            "id": 24237724,
            "node_id": "MDQ6VXNlcjI0MjM3NzI0",
            "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
            "gravatar_id": "",
            "url": "https://api.github.com/users/armory-jenkins",
            "html_url": "https://github.com/armory-jenkins",
            "followers_url": "https://api.github.com/users/armory-jenkins/followers",
            "following_url": "https://api.github.com/users/armory-jenkins/following{/other_user}",
            "gists_url": "https://api.github.com/users/armory-jenkins/gists{/gist_id}",
            "starred_url": "https://api.github.com/users/armory-jenkins/starred{/owner}{/repo}",
            "subscriptions_url": "https://api.github.com/users/armory-jenkins/subscriptions",
            "organizations_url": "https://api.github.com/users/armory-jenkins/orgs",
            "repos_url": "https://api.github.com/users/armory-jenkins/repos",
            "events_url": "https://api.github.com/users/armory-jenkins/events{/privacy}",
            "received_events_url": "https://api.github.com/users/armory-jenkins/received_events",
            "type": "User",
            "site_admin": false
        }
    }
]`,
			nil,
			"success",
			"No changes in dinghyfile.",
			0,
		},
		{
			"payload_does_not_have_status",
			`[
    {
        "url": "https://api.github.com/repos/armory-io/dinghy-test-jossue/statuses/613be8796790c0e88bd9e69054894b762ef16c88",
        "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
        "id": 10532473715,
        "node_id": "MDEzOlN0YXR1c0NvbnRleHQxMDUzMjQ3MzcxNQ==",
        "state": "failure",
        "description": null,
        "target_url": null,
        "context": "PR Title Checker (becomes a merge/squash commit)",
        "created_at": "2020-08-19T21:17:02Z",
        "updated_at": "2020-08-19T21:17:02Z",
        "creator": {
            "login": "armory-jenkins",
            "id": 24237724,
            "node_id": "MDQ6VXNlcjI0MjM3NzI0",
            "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
            "gravatar_id": "",
            "url": "https://api.github.com/users/armory-jenkins",
            "html_url": "https://github.com/armory-jenkins",
            "followers_url": "https://api.github.com/users/armory-jenkins/followers",
            "following_url": "https://api.github.com/users/armory-jenkins/following{/other_user}",
            "gists_url": "https://api.github.com/users/armory-jenkins/gists{/gist_id}",
            "starred_url": "https://api.github.com/users/armory-jenkins/starred{/owner}{/repo}",
            "subscriptions_url": "https://api.github.com/users/armory-jenkins/subscriptions",
            "organizations_url": "https://api.github.com/users/armory-jenkins/orgs",
            "repos_url": "https://api.github.com/users/armory-jenkins/repos",
            "events_url": "https://api.github.com/users/armory-jenkins/events{/privacy}",
            "received_events_url": "https://api.github.com/users/armory-jenkins/received_events",
            "type": "User",
            "site_admin": false
        }
    },
    {
        "url": "https://api.github.com/repos/armory-io/dinghy-test-jossue/statuses/613be8796790c0e88bd9e69054894b762ef16c88",
        "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
        "id": 10532473053,
        "node_id": "MDEzOlN0YXR1c0NvbnRleHQxMDUzMjQ3MzA1Mw==",
        "state": "pending",
        "description": null,
        "target_url": null,
        "context": "PR Title Checker (becomes a merge/squash commit)",
        "created_at": "2020-08-19T21:16:58Z",
        "updated_at": "2020-08-19T21:16:58Z",
        "creator": {
            "login": "armory-jenkins",
            "id": 24237724,
            "node_id": "MDQ6VXNlcjI0MjM3NzI0",
            "avatar_url": "https://avatars1.githubusercontent.com/u/24237724?v=4",
            "gravatar_id": "",
            "url": "https://api.github.com/users/armory-jenkins",
            "html_url": "https://github.com/armory-jenkins",
            "followers_url": "https://api.github.com/users/armory-jenkins/followers",
            "following_url": "https://api.github.com/users/armory-jenkins/following{/other_user}",
            "gists_url": "https://api.github.com/users/armory-jenkins/gists{/gist_id}",
            "starred_url": "https://api.github.com/users/armory-jenkins/starred{/owner}{/repo}",
            "subscriptions_url": "https://api.github.com/users/armory-jenkins/subscriptions",
            "organizations_url": "https://api.github.com/users/armory-jenkins/orgs",
            "repos_url": "https://api.github.com/users/armory-jenkins/repos",
            "events_url": "https://api.github.com/users/armory-jenkins/events{/privacy}",
            "received_events_url": "https://api.github.com/users/armory-jenkins/received_events",
            "type": "User",
            "site_admin": false
        }
    }
]`,
			nil,
			"",
			"",
			0,
		},
		{
			"request_fails",
			`dafds`,
			errors.New("invalid character 'd' looking for beginning of value"),
			"",
			"",
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := mock.NewMockFieldLogger(ctrl)
			logger.EXPECT().Error(gomock.Any()).Times(0)

			logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(tt.warning)

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tt.payload)
			}))
			defer ts.Close()
			// TODO: Do not use global variable. This will lead to side-effects.
			p := Push{
				Config: Config{
					Endpoint: ts.URL,
				},
				Repository: Repository{Organization: "armory", Name: "dinghy" },
				Logger: logger,
			}

			// This shouldn't throw exceptions/panics
			a, b, c := p.GetCommitStatus()
			var got = fmt.Sprintf("%v - %v - %v", a, b, c)
			var want = fmt.Sprintf("%v - %v - %v", tt.err, tt.status, tt.description)
			if got !=  want {
				t.Errorf("got %v and want %v", got, want )
				t.Fail()
			}
		})
	}

}