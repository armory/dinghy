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
	"errors"
	"fmt"
	"regexp"
)

// FileService serves a map[string]string of files -> file contents
type FileService map[string]string

// Download returns a file from the map
func (f FileService) Download(org, repo, file string) (string, error) {
	if ret, exists := f[file]; exists {
		return ret, nil
	}
	return "", errors.New("File not found")
}

// EncodeURL encodes a URL
func (f FileService) EncodeURL(org, repo, path string) string {
	return fmt.Sprintf(`%s/repos/%s/%s/contents/%s`, "https://github.com", org, repo, path)
}

// DecodeURL decodes a URL
func (f FileService) DecodeURL(url string) (org, repo, path string) {
	targetExpression := ".*/repos/(.+)/(.+)/contents/(.+)"
	r, _ := regexp.Compile(targetExpression)
	match := r.FindStringSubmatch(url)
	org = match[1]
	repo = match[2]
	path = match[3]
	return
}