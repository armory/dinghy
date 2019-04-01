package github

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/armory-io/dinghy/pkg/cache/local"
	"github.com/armory-io/dinghy/pkg/util"

	"github.com/google/go-github/github"

	log "github.com/sirupsen/logrus"
)

// FileService is for working with repositories
type FileService struct {
	cache          local.Cache
	GitHubToken    string
	GitHubEndpoint string
}

// Download a file from github
// note that "path" is the full path relative to the repo root
// eg: src/foo/bar/filename
func (f *FileService) Download(org, repo, path string) (string, error) {
	url := f.EncodeURL(org, repo, path)
	body := f.cache.Get(url)
	if body != "" {
		return body, nil
	}

	ctx := context.Background()
	client, err := newGitHubClient(ctx, f.GitHubEndpoint, f.GitHubToken)
	if err != nil {
		log.Errorf("unable to create Github client for Download request: %s", err)
		return "", errors.New("unable to create Github client for Download request")
	}

	r, err := client.Repositories.DownloadContents(ctx, org, repo, path, nil)
	if err != nil {
		if e, ok := err.(*github.RateLimitError); ok {
			return "", &util.GithubRateLimitErr{RateLimit: e.Rate.Limit, RateReset: e.Rate.Reset.String()}
		}
		return "", err
	}

	b := new(bytes.Buffer)
	b.ReadFrom(r)
	r.Close()

	f.cache.Add(url, b.String())

	return b.String(), nil
}

// EncodeURL returns the git url for a given org, repo, path
func (f *FileService) EncodeURL(org, repo, path string) string {
	// this is only used for caching purposes
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
