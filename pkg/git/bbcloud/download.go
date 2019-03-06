package bbcloud

import (
	"errors"
	"fmt"
	"github.com/armory-io/dinghy/pkg/cache/local"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"regexp"
)

// FileService is for working with repositories
type FileService struct {
	cache           local.Cache
	BbcloudEndpoint string
	BbcloudToken    string
	BbcloudUsername string
}

// Download downloads a file from Bitbucket Cloud.
// The API returns the file's contents as a paginated list of lines
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
	req.SetBasicAuth(f.BbcloudUsername, f.BbcloudToken)
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Errorf("Error downloading file from %s: Status: %d", url, resp.StatusCode)
		err = errors.New("Download error")
		return "", err
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	retString := string(ret)

	f.cache.Add(url, retString)
	return retString, nil
}

// EncodeURL returns the git url for a given org, repo, path
func (f *FileService) EncodeURL(org, repo, path string) string {
	return fmt.Sprintf(`%s/repositories/%s/%s/src/master/%s?raw`, f.BbcloudEndpoint, org, repo, path)
}

// DecodeURL takes a url and returns the org, repo, path
func (f *FileService) DecodeURL(url string) (org, repo, path string) {
	r, _ := regexp.Compile(`/repositories/(.+)/(.+)/src/master/(.+)\?raw`)
	match := r.FindStringSubmatch(url)
	org = match[1]
	repo = match[2]
	path = match[3]
	return
}
