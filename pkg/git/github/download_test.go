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

	"github.com/stretchr/testify/assert"
)

func TestEncodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		expected string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			expected: "https://api.github.com/repos/armory/armory/contents/my/path.yml",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			expected: "https://mygithub.com/repos/armory/armory/contents/my/path.yml",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		if out := downloader.EncodeURL(c.owner, c.repo, c.path); out != c.expected {
			t.Errorf("%s did not match expected %s", out, c.expected)
		}
	}
}

func TestDecodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		url      string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			url:      "https://api.github.com/repos/armory/armory/contents/my/path.yml",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			url:      "https://mygithub.com/repos/armory/armory/contents/my/path.yml",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		org, repo, path := downloader.DecodeURL(c.url)

		if org != c.owner {
			t.Errorf("%s did not match expected owner %s", org, c.owner)
		}

		if repo != c.repo {
			t.Errorf("%s did not match expected repo %s", repo, c.repo)
		}

		if path != c.path {
			t.Errorf("%s did not match expected path %s", path, c.path)
		}
	}
}

func TestDownload(t *testing.T) {
	testCases := map[string]struct {
		org         string
		repo        string
		path        string
		fs          *FileService
		expected    string
		expectedErr error
	}{
		"success": {
			org:  "org",
			repo: "repo",
			path: "path",
			fs: &FileService{
				GitHub: &GitHubTest{contents: "file contents"},
			},
			expected:    "file contents",
			expectedErr: nil,
		},
		"error": {
			org:  "org",
			repo: "repo",
			path: "path",
			fs: &FileService{
				GitHub: &GitHubTest{
					contents: "",
					err:      errors.New("fail"),
				},
			},
			expected:    "",
			expectedErr: errors.New("fail"),
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual, err := tc.fs.Download(tc.org, tc.repo, tc.path)
			assert.Equal(t, tc.expected, actual)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}

			// test caching
			v := tc.fs.cache.Get(tc.fs.EncodeURL("org", "repo", "path"))
			assert.Equal(t, tc.expected, v)
		})
	}
}
