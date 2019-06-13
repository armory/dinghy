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

package stash

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDecodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		url      string
		branch   string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "myrepo",
			path:     "my/path.yml",
			url:      "/projects/armory/repos/myrepo/browse/my/path.yml?raw",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			StashEndpoint: c.endpoint,
		}
		org, repo, path, branch := downloader.DecodeURL(c.url)

		assert.Equal(t, c.owner, org)
		assert.Equal(t, c.repo, repo)
		assert.Equal(t, c.path, path)
		assert.Equal(t, c.branch, branch)
	}
}
