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

// Push contains data about a push full of commits
type Push struct {
	Payload         WebhookPayload
	ChangedFiles    []string
	BbcloudEndpoint string
	BBcloudUsername string
	BbcloudToken    string
}

// WebhookPayload is the payload from the webhook
type WebhookPayload struct {
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Project  struct {
			Key string `json:"key"`
		} `json:"project"`
	} `json:"repository"`

	Push struct {
		Changes []WebhookChange `json:"changes"`
	}
}

// WebhookChange is a change object from the webhook
type WebhookChange struct {
	New struct {
		Name   string `json:"name"`
		Target struct {
			Hash string `json:"hash"`
		}
	}
	Old struct {
		Target struct {
			Hash string `json:"hash"`
		}
	}
}

// IsMaster detects if a change was on master
func (c *WebhookChange) IsMaster() bool {
	return c.New.Name == "master"
}

// APIResponse is the response from Stash API
type APIResponse struct {
	PagedAPIResponse
	Diffs []APIDiff `json:"values"`
}

// APIDiff is a diff returned by the Stash API
type APIDiff struct {
	New struct {
		Path string `json:"path"`
	} `json:"new"`
}

func (p *Push) getFilesChanged(fromCommitHash, toCommitHash string, page int) (nextPage int, err error) {
	url := fmt.Sprintf(
		`%s/repositories/%s/diffstat/%s`,
		p.BbcloudEndpoint,
		p.Payload.Repository.FullName,
		toCommitHash,
	)
	log.Debug("ApiCall: ", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return page, err
	}

	query := req.URL.Query()
	query.Add("since", fromCommitHash)
	query.Add("withComments", "false")
	if page != 1 {
		query.Add("page", strconv.Itoa(page))
	}
	req.URL.RawQuery = query.Encode()
	log.Debugf("Raw query: %s\n", req.URL.RawQuery)
	req.SetBasicAuth(p.BBcloudUsername, p.BbcloudToken)

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Error(err)
		return page, err
	}

	var body APIResponse
	respRaw, err := ioutil.ReadAll(resp.Body)
	respString := string(respRaw)
	log.Debugf("APIResponse: %s\n", respString)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Errorf("Error response from server: status code: %d\n", resp.StatusCode)
		return page, err
	}

	err = json.Unmarshal(respRaw, &body)
	log.Debugf("APIResponse decoded: %v\n", body)
	if err != nil {
		return page, err
	}
	if body.CurrentPage < body.NumberOfPages {
		nextPage = body.CurrentPage + 1
	}
	for _, diff := range body.Diffs {
		p.ChangedFiles = append(p.ChangedFiles, diff.New.Path)
	}

	return page, nil
}

type BbCloudConfig struct {
	Username string
	Token    string
	Endpoint string
}

// NewPush creates a new Push
func NewPush(payload WebhookPayload, cfg BbCloudConfig) (*Push, error) {
	p := &Push{
		Payload:         payload,
		ChangedFiles:    make([]string, 0),
		BbcloudEndpoint: cfg.Endpoint,
		BbcloudToken:    cfg.Token,
		BBcloudUsername: cfg.Username,
	}

	for _, change := range p.changes() {
		if !change.IsMaster() {
			continue
		}
		for page := 1; true; {
			nextPage, err := p.getFilesChanged(change.Old.Target.Hash, change.New.Target.Hash, page)
			if err != nil {
				return nil, err
			}
			if page == nextPage {
				break
			}
		}
	}

	return p, nil
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
