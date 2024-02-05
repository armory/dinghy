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
	"github.com/armory/dinghy/pkg/log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/armory/dinghy/pkg/cache/local"
)

// FileService is for working with repositories
type FileService struct {
	cache  local.Cache
	Config Config
	Logger log.DinghyLog
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
	req.SetBasicAuth(f.Config.Username, f.Config.Token)
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

	var err error

	if file, err := f.getFile(org, repo, path, branch); err == nil {
		return file, nil
	}

	// If we are unable to retrieve files from the designated branch, we should attempt to access an alternative branch.
	// It is not always clear which branch is the true master branch due to different naming conventions.
	// Therefore, we will need to try one of them [main, master] to ensure we can access the necessary files.
	var branchAlternatives = map[string]string{
		"master": "main",
		"main":   "master",
	}

	if alternativeBranchName, ok := branchAlternatives[branch]; ok {
		f.Logger.Info(fmt.Sprintf("DownloadContents failed with %v branch, trying with %v branch", branch, alternativeBranchName))

		if altFile, altErr := f.getFile(org, repo, path, alternativeBranchName); altErr == nil {
			f.Logger.Infof("Download from secondary branch %v succeeded", alternativeBranchName)
			return altFile, nil
		}

		f.Logger.Errorf("Download failed also for branch %v", alternativeBranchName)
		// We deliberately want to return the original err, not altErr
		return "", err
	}

	return "", nil
}

// EncodeURL returns the git url for a given org, repo, path and branch
func (f *FileService) EncodeURL(org, repo, path, branch string) string {
	return fmt.Sprintf(`%s/projects/%s/repos/%s/browse/%s?at=%s&raw`, f.Config.Endpoint, org, repo, path, branch)
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

func (f *FileService) getFile(org, repo, path, branch string) (string, error) {
	url := f.EncodeURL(org, repo, path, branch)
	body := f.cache.Get(url)
	if body != "" {
		f.Logger.Infof(fmt.Sprintf("Found cached file for key: %v\n", url))
		return body, nil
	}

	f.Logger.Infof(fmt.Sprintf("Didn't find any cached files for key: %v. Will proceed to download\n", url))

	result, err := f.getAllLines(url)

	if err != nil {
		f.Logger.Errorf(fmt.Sprintf("Failed to download file from: %v\n", url))
		return "", err
	}

	f.Logger.Debug(fmt.Sprintf("Successfully downloaded file from %v branch", branch))
	f.cache.Add(url, result)
	return result, nil
}

func (f *FileService) getAllLines(url string) (string, error) {
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
	return strings.Join(allLines, "\n"), nil
}
