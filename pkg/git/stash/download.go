package stash

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory-io/dinghy/pkg/cache/local"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/armory-io/dinghy/pkg/settings"
	log "github.com/sirupsen/logrus"
)

// FileService is for working with repositories
type FileService struct {
	cache local.Cache
}

// FileContentsResponse contains response from Stash when you fetch a file
type FileContentsResponse struct {
	PagedAPIResponse
	Lines []map[string]string `json:"lines"`
}

func (f *FileService) downloadLines(url string, start int) (lines []string, nextStart int, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	if start != -1 {
		query := req.URL.Query()
		query.Add("start", strconv.Itoa(start))
		req.URL.RawQuery = query.Encode()
	}
	req.SetBasicAuth(settings.S.StashUsername, settings.S.StashToken)
	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		log.Errorf("Error downloading file from %s: Stauts: %d", url, resp.StatusCode)
		err = errors.New("Download error")
		return
	}

	var body FileContentsResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return
	}
	if !body.IsLastPage {
		if body.NextPageStart == 0 {
			log.Errorf("IsLastPage is false but NextPageStart is nil for: %s", url)
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
func (f *FileService) Download(org, repo, path string) (string, error) {
	url := f.EncodeURL(org, repo, path)
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

// EncodeURL returns the git url for a given org, repo, path
func (f *FileService) EncodeURL(org, repo, path string) string {
	return fmt.Sprintf(`%s/projects/%s/repos/%s/browse/%s?raw`, settings.S.StashEndpoint, org, repo, path)
}

// DecodeURL takes a url and returns the org, repo, path
func (f *FileService) DecodeURL(url string) (org, repo, path string) {
	r, _ := regexp.Compile(`/projects/(.+)/repos/(.+)/browse/(.+)\?raw`)
	match := r.FindStringSubmatch(url)
	org = match[1]
	repo = match[2]
	path = match[3]
	return
}
