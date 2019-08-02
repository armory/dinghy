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
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/armory/dinghy/pkg/cache/local"

	"github.com/sirupsen/logrus"
)

// FileService is for working with repositories
type FileService struct {
	cache         local.Cache
	StashEndpoint string
	StashToken    string
	StashUsername string
	Logger        logrus.FieldLogger
}

// FileContentsResponse contains response from Stash when you fetch a file
type FileContentsResponse struct {
	PagedAPIResponse
	Lines []map[string]string `json:"lines"`
}

func (f *FileService) downloadLines(url string, start int) (lines []string, nextStart int, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		f.Logger.Warnf("downloadLines: NewRequest for %s failed: %s", url, err.Error())
		return
	}
	if start != -1 {
		query := req.URL.Query()
		query.Add("start", strconv.Itoa(start))
		req.URL.RawQuery = query.Encode()
	}
	req.SetBasicAuth(f.StashUsername, f.StashToken)
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		f.Logger.Warnf("downloadLines: DefaultClient.Do for %s failed: %s", req.URL.RawQuery, err.Error())
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Error downloading file from %s: Status: %d", url, resp.StatusCode)
		return
	}

	var body FileContentsResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		f.Logger.Warnf("JSON parse error downloading Stash file from %s, got: %s", url, resp.Body)
		return
	}
	if !body.IsLastPage {
		if body.NextPageStart == 0 {
			f.Logger.Errorf("IsLastPage is false but NextPageStart is nil for: %s", url)
		} else {
			nextStart = body.NextPageStart
		}
	}
	for _, line := range body.Lines {
		lines = append(lines, line["text"])
	}
	return
}

// Download downloads a file from Stash.
// Stash's API returns the file's contents as a paginated list of lines
func (f *FileService) Download(org, repo, path, branch string) (string, error) {
	// bitbucket server & stash require the full path to the branch name
	branch = fmt.Sprintf("refs/heads/%s", branch)

	url := f.EncodeURL(org, repo, path, branch)
	body := f.cache.Get(url)
	if body != "" {
		return body, nil
	}
	allLines := make([]string, 0)
	start := -1
	for start != 0 {
		lines, nextStart, err := f.downloadLines(url, start)
		if err != nil {
			return "", err
		}
		allLines = append(allLines, lines...)
		start = nextStart
	}
	ret := strings.Join(allLines, "\n")
	f.cache.Add(url, ret)
	return ret, nil
}

// EncodeURL returns the git url for a given org, repo, path and branch
func (f *FileService) EncodeURL(org, repo, path, branch string) string {
	return fmt.Sprintf(`%s/projects/%s/repos/%s/browse/%s?at=%s&raw`, f.StashEndpoint, org, repo, path, branch)
}

// DecodeURL takes a url and returns the org, repo, path and branch
func (f *FileService) DecodeURL(url string) (org, repo, path, branch string) {
	r, _ := regexp.Compile(`/projects/(.+)/repos/(.+)/browse/(.+)\?at=(.+)\&raw`)
	match := r.FindStringSubmatch(url)
	org = match[1]
	repo = match[2]
	path = match[3]
	branch = match[4]
	return
}
