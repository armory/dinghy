package github

import (
	"fmt"
	"github.com/armory-io/dinghy/pkg/settings"
	"io/ioutil"
	"net/http"
	"regexp"
)

// FileService is for working with repositories
type FileService struct{}

// Download a file from github.
// note that "path" is the full path relative to the repo root
// eg: src/foo/bar/filename
func (f *FileService) Download(org, repo, path string) (string, error) {
	url := f.GitURL(org, repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "token "+settings.S.GitHubToken)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GitURL returns the git url for a given org, repo, path
func (f *FileService) GitURL(org, repo, path string) string {
	return fmt.Sprintf(`https://api.github.com/repos/%s/%s/contents/%s`, org, repo, path)
}

// ParseGitURL takes a url and returns the org, repo, path
func (f *FileService) ParseGitURL(url string) (org, repo, path string) {
	r, _ := regexp.Compile("https://api.github.com/repos/(.+)/(.+)/contents/(.+)")
	match := r.FindStringSubmatch(url)
	org = match[1]
	repo = match[2]
	path = match[3]
	return
}
