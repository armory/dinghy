package github

import (
	"fmt"
	"github.com/armory-io/dinghy/pkg/settings"
	"io/ioutil"
	"net/http"
)

// FileService is for working with repositories
type FileService struct{}

// Download a file from github.
// note that "path" is the full path relative to the repo root
// eg: src/foo/bar/filename
func (f *FileService) Download(org, repo, path string) (string, error) {
	url := fmt.Sprintf(`https://api.github.com/repos/%s/%s/contents/%s`, org, repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "token "+settings.GitHubToken)
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
