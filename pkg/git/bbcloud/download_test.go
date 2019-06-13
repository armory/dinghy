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

package bbcloud

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testData struct {
	endpoint string
	owner    string
	repo     string
	path     string
	branch   string
	url      string
}

func TestDecodeUrl(t *testing.T) {
	cases := []testData{
		testData{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "myrepo",
			path:     "my/path.yml",
			branch:   "master",
			url:      "/repositories/armory/myrepo/src/master/my/path.yml?raw",
		},
		testData{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "myrepo",
			path:     "my/path.yml",
			branch:   "prod",
			url:      "/repositories/armory/myrepo/src/prod/my/path.yml?raw",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			BbcloudEndpoint: c.endpoint,
		}
		org, repo, path, branch := downloader.DecodeURL(c.url)

		assert.Equal(t, c.owner, org)
		assert.Equal(t, c.repo, repo)
		assert.Equal(t, c.path, path)
		assert.Equal(t, c.branch, branch)
	}
}

func TestEncodeUrl(t *testing.T) {
	cases := []testData{
		testData{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "myrepo",
			path:     "my/path.yml",
			branch:   "master",
			url:      "https://api.github.com/repositories/armory/myrepo/src/master/my/path.yml?raw",
		},
		testData{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "myrepo",
			path:     "my/path.yml",
			branch:   "prod",
			url:      "https://api.github.com/repositories/armory/myrepo/src/prod/my/path.yml?raw",
		},
		testData{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "myrepo",
			path:     "my/path.yml",
			url:      "https://api.github.com/repositories/armory/myrepo/src/master/my/path.yml?raw",
			// leaving out branch, should get default "" value.
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			BbcloudEndpoint: c.endpoint,
		}
		url := downloader.EncodeURL(c.owner, c.repo, c.path, c.branch)
		assert.Equal(t, c.url, url)
	}
}
