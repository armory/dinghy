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
	"fmt"
	"github.com/armory/dinghy/pkg/log"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/armory/dinghy/pkg/cache/local"
)

// FileService is for working with repositories
type FileService struct {
	cache  local.Cache
	Config Config
	Logger log.DinghyLog
}

// Download downloads a file from Bitbucket Cloud.
// The API returns the file's contents as a paginated list of lines
func (f *FileService) Download(org, repo, path, branch string) (string, error) {
	url := f.EncodeURL(org, repo, path, branch)
	body := f.cache.Get(url)
	if body != "" {
		return body, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(f.Config.Username, f.Config.Token)
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error downloading file from %s: Status: %d", url, resp.StatusCode)
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	retString := string(ret)

	f.cache.Add(url, retString)
	return retString, nil
}

// EncodeURL returns the git url for a given org, repo, path and branch
func (f *FileService) EncodeURL(org, repo, path, branch string) string {
	return fmt.Sprintf(`%s/repositories/%s/%s/src/%s/%s?raw`, f.Config.Endpoint, org, repo, branch, path)
}

// DecodeURL takes a url and returns the org, repo, path and branch
func (f *FileService) DecodeURL(url string) (org, repo, path, branch string) {
	r, _ := regexp.Compile(`/repositories/(.+)/(.+)/src/([^/]+)/(.+)\?raw`)
	match := r.FindStringSubmatch(url)
	org = match[1]
	repo = match[2]
	branch = match[3]
	path = match[4]
	return
}
