package github

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/armory-io/dinghy/pkg/cache/local"

	log "github.com/sirupsen/logrus"
)

// FileService is for working with repositories
type FileService struct {
	cache          local.Cache
	GitHubToken    string
	GitHubEndpoint string
}

// Download a file from github.
// note that "path" is the full path relative to the repo root
// eg: src/foo/bar/filename
func (f *FileService) Download(org, repo, path string) (string, error) {
	url := f.EncodeURL(org, repo, path)
	body := f.cache.Get(url)
	if body != "" {
		return body, nil
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "token "+f.GitHubToken)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Errorf("Error downloading file from %s: Status: %d", url, resp.StatusCode)
		return "", errors.New("Download error")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	f.cache.Add(url, string(b))

	// log the current rate limit
	rateLimit, err := getRateLimit(resp.Header)
	if err != nil {
		log.Debugf("Error retrieving rate limit header: %s", err)
	}
	log.Debugf("Current rate limit: %s", rateLimit)

	return string(b), nil
}

// EncodeURL returns the git url for a given org, repo, path
func (f *FileService) EncodeURL(org, repo, path string) string {
	return fmt.Sprintf(`%s/repos/%s/%s/contents/%s`, f.GitHubEndpoint, org, repo, path)
}

// DecodeURL takes a url and returns the org, repo, path
func (f *FileService) DecodeURL(url string) (org, repo, path string) {
	targetExpression := fmt.Sprintf("%s/repos/(.+)/(.+)/contents/(.+)", f.GitHubEndpoint)
	r, _ := regexp.Compile(targetExpression)
	match := r.FindStringSubmatch(url)
	org = match[1]
	repo = match[2]
	path = match[3]
	return
}
