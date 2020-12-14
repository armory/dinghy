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
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"io/ioutil"
	"net/http"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewRoundTripTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func NewTestClient(statusCode int, contents string) *gitlab.Client {
	httpClient := NewRoundTripTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewBufferString(contents)),
			Header:     make(http.Header),
		}
	})

	return gitlab.NewClient(httpClient, "token")
}

func TestEncodeUrl(t *testing.T) {
	client := NewTestClient(200, "")

	testCases := map[string]struct {
		owner    string
		repo     string
		path     string
		branch   string
		expected string
	}{
		"success": {
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			expected: fmt.Sprintf("%sprojects/armory/armory/repository/files/my/path.yml?ref=mybranch", client.BaseURL()),
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			downloader := &FileService{
				Client: client,
			}
			actual := downloader.EncodeURL(tc.owner, tc.repo, tc.path, tc.branch)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestDecodeUrl(t *testing.T) {
	client := NewTestClient(200, "")

	testCases := map[string]struct {
		owner  string
		repo   string
		path   string
		branch string
		url    string
	}{
		"success": {
			owner:  "armory",
			repo:   "armory",
			path:   "my/path.yml",
			branch: "mybranch",
			url:    fmt.Sprintf("%sprojects/armory/armory/repository/files/my/path.yml?ref=mybranch", client.BaseURL()),
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			downloader := &FileService{
				Client: client,
			}
			org, repo, path, branch := downloader.DecodeURL(tc.url)
			assert.Equal(t, tc.owner, org)
			assert.Equal(t, tc.repo, repo)
			assert.Equal(t, tc.path, path)
			assert.Equal(t, tc.branch, branch)
		})
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
				Client: NewTestClient(200, "file contents"),
			},
			expected:    "file contents",
			expectedErr: nil,
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
			v := tc.fs.cache.Get(tc.fs.EncodeURL(tc.org, tc.repo, tc.path, tc.branch))
			assert.Equal(t, tc.expected, v)
		})
	}
}
