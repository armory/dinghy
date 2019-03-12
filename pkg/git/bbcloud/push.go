package bbcloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/armory-io/dinghy/pkg/git"
	log "github.com/sirupsen/logrus"
)

// -----------------------------------------------------------------------------
// Bitbucket Cloud data types
// -----------------------------------------------------------------------------

// Payload received in a "push" type webhook
type WebhookPayload struct {
	Repository WebhookRepository `json:"repository"`
	Push       WebhookPush       `json:"push"`
}

type WebhookProject struct {
	Key string `json:"key"`
}

type WebhookRepository struct {
	Name     string         `json:"name"`
	FullName string         `json:"full_name"`
	Project  WebhookProject `json:"project"`
}

type WebhookPush struct {
	Changes []WebhookChange `json:"changes"`
}

// Change details in a webhook payload
type WebhookChange struct {
	New WebhookChangeComparison `json:"new"`
	Old WebhookChangeComparison `json:"old"`
}

type WebhookChangeComparison struct {
	Name   string              `json:"name"`
	Target WebhookChangeTarget `json:"target"`
}

type WebhookChangeTarget struct {
	Hash string `json:"hash"`
}

// Response of getting diff details of a commit. Documentation:
// https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Busername%7D/%7Brepo_slug%7D/diffstat/%7Bspec%7D
type DiffStatResponse struct {
	PagedAPIResponse
	Diffs []APIDiff `json:"values"`
}

// Details of a single file changed
type APIDiff struct {
	New struct {
		Path string `json:"path"`
	} `json:"new"`
}

// -----------------------------------------------------------------------------
// Dinghy data types
// -----------------------------------------------------------------------------

// Connection details for talking with bitbucket cloud
type Config struct {
	Username string
	Token    string
	Endpoint string
}

// Information about what files changed in a push
type Push struct {
	Payload      WebhookPayload
	ChangedFiles []string
}

// -----------------------------------------------------------------------------
// Functions implementation
// -----------------------------------------------------------------------------

// Converts the raw webhook payload sent by Bitbucket Cloud, to an internal Push structure
// with the list of files changed.
func NewPush(payload WebhookPayload, cfg Config) (*Push, error) {
	p := &Push{
		Payload:      payload,
		ChangedFiles: make([]string, 0),
	}

	changedFilesMap := map[string]bool{}

	for _, change := range p.changes() {
		// Only process changes in "master" branch
		if !change.IsMaster() {
			continue
		}
		for page := 1; true; page++ {
			changedFiles, nextPage, err := getFilesChanged(change.Old.Target.Hash, change.New.Target.Hash, page, cfg,
				payload.Repository.FullName)
			if err != nil {
				return nil, err
			}
			// trick for removing duplicates from a list (keys are unique)
			for _, file := range changedFiles {
				changedFilesMap[file] = true
			}
			if page == nextPage {
				break
			}
		}
	}

	for file := range changedFilesMap {
		p.ChangedFiles = append(p.ChangedFiles, file)
	}

	return p, nil
}

// Find in the webook payload if the change was done in "master" branch
func (c *WebhookChange) IsMaster() bool {
	return c.New.Name == "master"
}

func getFilesChanged(fromCommitHash, toCommitHash string, page int, cfg Config,
	repoName string) (changedFiles []string, nextPage int, err error) {

	url := fmt.Sprintf(
		`%s/repositories/%s/diffstat/%s`,
		cfg.Endpoint,
		repoName,
		toCommitHash,
	)

	nextPage = page

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []string{}, page, err
	}

	query := req.URL.Query()
	query.Add("since", fromCommitHash)
	query.Add("withComments", "false")
	if page > 1 {
		query.Add("page", strconv.Itoa(page))
	}
	req.URL.RawQuery = query.Encode()
	log.Infof("Diffstat request: GET %s?%s\n", url, req.URL.RawQuery)
	req.SetBasicAuth(cfg.Username, cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	changedFiles, hasNext, err := handleDiffstatResponse(resp, err)
	if hasNext {
		nextPage = page + 1
	}

	return
}

func handleDiffstatResponse(resp *http.Response, err error) (changedFiles []string, hasNext bool, outErr error) {
	if err != nil {
		log.Errorf("Diffstat error: %v", err)
		return []string{}, false, err
	}

	var apiResponse DiffStatResponse
	respRaw, err := ioutil.ReadAll(resp.Body)
	respString := string(respRaw)
	log.Debugf("DiffStatResponse: %s\n", respString)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Errorf("Diffstat error: response status code %d\n", resp.StatusCode)
		return []string{}, false, err
	}

	err = json.Unmarshal(respRaw, &apiResponse)
	if err != nil {
		return []string{}, false, err
	}

	if apiResponse.CurrentPage < apiResponse.NumberOfPages {
		hasNext = true
	} else {
		hasNext = false
	}

	for _, diff := range apiResponse.Diffs {
		changedFiles = append(changedFiles, diff.New.Path)
	}

	return changedFiles, hasNext, nil
}

// ContainsFile checks to see if a given file is in the push.
func (p *Push) ContainsFile(file string) bool {
	for _, name := range p.ChangedFiles {
		if name == file {
			return true
		}
	}
	return false
}

// Files returns a slice containing filenames that were added/modified
func (p *Push) Files() []string {
	return p.ChangedFiles
}

// Repo returns the name of the repo.
func (p *Push) Repo() string {
	return p.Payload.Repository.Name
}

// Org returns the name of the project.
func (p *Push) Org() string {
	parts := strings.Split(p.Payload.Repository.FullName, "/")
	if parts != nil && len(parts) > 0 {
		return parts[0]
	} else {
		return ""
	}
}

func (p *Push) changes() []WebhookChange {
	return p.Payload.Push.Changes
}

// IsMaster detects if the branch is master.
func (p *Push) IsMaster() bool {
	for _, change := range p.changes() {
		if change.IsMaster() {
			return true
		}
	}
	return false
}

// SetCommitStatus sets a commit status
func (p *Push) SetCommitStatus(s git.Status) {
}
