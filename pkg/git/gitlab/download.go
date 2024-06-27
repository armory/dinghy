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
	"fmt"
	"github.com/armory/dinghy/pkg/cache/local"
	"github.com/armory/dinghy/pkg/log"
	gitlab "github.com/xanzy/go-gitlab"
	"regexp"
	"strings"
)

type GitLabClient interface {
	DownloadContents(string, string, string, string) (string, error)
	// CreateStatus(*Status, string, string, string) error
	GetEndpoint() string
}

// FileService is for working with repositories
type FileService struct {
	cache  local.Cache
	Client *gitlab.Client
	Logger log.DinghyLog
}

// eDownload a file from gitlab
// note that "path" is the full path relative to the repo root
// eg: src/foo/bar/filename
func (f *FileService) Download(org, repo, path, branch string) (string, error) {

	// Note:  We're really only using EncodeURL here for caching.  The go-gitlab
	// client knows how to create the URL internally.
	branch = strings.Replace(branch, "refs/heads/", "", 1)
	url := f.EncodeURL(org, repo, path, branch)
	body := f.cache.Get(url)
	if body != "" {
		return body, nil
	}

	// go-gitlab uses a "pid" which is the org/repo combo. Done here for
	// clarity
	pid := fmt.Sprintf("%s/%s", org, repo)

	contents, _, err := f.Client.RepositoryFiles.GetRawFile(pid, path, &gitlab.GetRawFileOptions{Ref: &branch})
	if err != nil {
		f.Logger.Error(err)
		return "", err
	}

	str := string(contents)
	f.cache.Add(url, str)

	return str, nil
}

// EncodeURL returns the git url for a given org, repo, path and branch
func (f *FileService) EncodeURL(org, repo, path, branch string) string {
	// this is only used for caching purposes
	return fmt.Sprintf(`%sprojects/%s/%s/repository/files/%s?ref=%s`, f.Client.BaseURL(), org, repo, path, branch)
}

// DecodeURL takes a url and returns the org, repo, path and branch
func (f *FileService) DecodeURL(url string) (org, repo, path, branch string) {
	targetExpression := fmt.Sprintf(`%sprojects/(.+)/(.+)/repository/files/(.+)\?ref=(.+)`, f.Client.BaseURL())
	r, _ := regexp.Compile(targetExpression)
	match := r.FindStringSubmatch(url)
	org = match[1]
	repo = match[2]
	path = match[3]
	branch = match[4]
	return
}
