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
	"bytes"
	"errors"
	"github.com/armory/dinghy/pkg/log"
	"testing"

	_ "github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestEncodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		branch   string
		expected string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			expected: "https://api.github.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			expected: "https://mygithub.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		actual := downloader.EncodeURL(c.owner, c.repo, c.path, c.branch)
		assert.Equal(t, c.expected, actual)
	}
}

func TestEncodeUrlWithLeadingSlashs(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		branch   string
		expected string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "/my/path.yml",
			branch:   "mybranch",
			expected: "https://api.github.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			expected: "https://mygithub.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		actual := downloader.EncodeURL(c.owner, c.repo, c.path, c.branch)
		assert.Equal(t, c.expected, actual)
	}
}

func TestDecodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		branch   string
		url      string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			url:      "https://api.github.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			url:      "https://mygithub.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		org, repo, path, branch := downloader.DecodeURL(c.url)
		assert.Equal(t, c.owner, org)
		assert.Equal(t, c.repo, repo)
		assert.Equal(t, c.path, path)
		assert.Equal(t, c.branch, branch)
	}
}

func TestDownload(t *testing.T) {
	testCases := map[string]struct {
		org         string
		repo        string
		path        string
		branch      string
		fs          *FileService
		expected    string
		expectedErr error
	}{
		"success": {
			org:    "org",
			repo:   "repo",
			path:   "path",
			branch: "branch",
			fs: &FileService{
				GitHub: &GitHubTest{contents: "file contents"},
				Logger: log.DinghyLogs{Logs: map[string]log.DinghyLogStruct{
					log.SystemLogKey : {
						Logger:         logrus.New(),
						LogEventBuffer: &bytes.Buffer{},
					},
				}},
			},
			expected:    "file contents",
			expectedErr: nil,
		},
		"error": {
			org:    "org",
			repo:   "repo",
			path:   "path",
			branch: "branch",
			fs: &FileService{
				GitHub: &GitHubTest{
					contents: "",
					err:      errors.New("fail"),
				},
				Logger: log.DinghyLogs{Logs: map[string]log.DinghyLogStruct{
					log.SystemLogKey : {
						Logger:         logrus.New(),
						LogEventBuffer: &bytes.Buffer{},
					},
				}},
			},
			expected:    "",
			expectedErr: errors.New("fail"),
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual, err := tc.fs.Download(tc.org, tc.repo, tc.path, tc.branch)
			assert.Equal(t, tc.expected, actual)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}

			// test caching
			v := tc.fs.cache.Get(tc.fs.EncodeURL("org", "repo", "path", "branch"))
			assert.Equal(t, tc.expected, v)
		})
	}
}
