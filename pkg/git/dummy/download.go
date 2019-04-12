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

package dummy

import (
	"fmt"

	"github.com/armory-io/dinghy/pkg/git/github"
)

// FileService serves a map[string]string of files -> file contents
type FileService map[string]string

// Download returns a file from the map
func (f FileService) Download(org, repo, file string) (string, error) {
	if ret, exists := f[file]; exists {
		return ret, nil
	}
	return "", nil
}

// EncodeURL encodes a URL
func (f FileService) EncodeURL(org, repo, path string) string {
	return fmt.Sprintf(`%s/repos/%s/%s/contents/%s`, "https://github.com", org, repo, path)
}

// DecodeURL decodes a URL
func (f FileService) DecodeURL(url string) (org, repo, path string) {
	return (&github.FileService{}).DecodeURL(url)
}
