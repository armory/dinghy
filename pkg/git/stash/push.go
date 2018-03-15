package stash

import (
	"encoding/json"
	"fmt"
	"net/http"

	"strconv"

	"github.com/armory-io/dinghy/pkg/settings"
	log "github.com/sirupsen/logrus"
)

// Push contains data about a push full of commits
type Push struct {
	Payload      WebhookPayload
	ChangedFiles []string
}

// WebhookPayload is the payload from the webhook
type WebhookPayload struct {
	EventKey   string            `json:"eventKey"`
	Repository WebhookRepository `json:"repository"`
	Changes    []WebhookChange   `json:"changes"`
}

// WebhookRepository is a repo object from the webhook
type WebhookRepository struct {
	Slug    string         `json:"slug"`
	Project WebhookProject `json:"project"`
}

// WebhookProject is a project object from the webhook
type WebhookProject struct {
	Key string `json:"key"`
}

// WebhookChange is a change object from the webhook
type WebhookChange struct {
	RefID    string `json:"refId"`
	FromHash string `json:"fromHash"`
	ToHash   string `json:"toHash"`
}

// APIResponse is the response from Stash API
type APIResponse struct {
	PagedAPIResponse
	Diffs []APIDiff `json:"values"`
}

// APIDiff is a diff returned by the Stash API
type APIDiff struct {
	Destination struct {
		Path string `json:"toString"`
	} `json:"path"`
}

// IsMaster detects if a change was on master
func (c *WebhookChange) IsMaster() bool {
	return c.RefID == "refs/heads/master"
}

func getFilesChanged(push *Push, fromCommitHash, toCommitHash string, start int) (nextStart int, err error) {
	url := fmt.Sprintf(
		`%s/projects/%s/repos/%s/commits/%s/changes`,
		settings.StashEndpoint,
		push.Payload.Repository.Project.Key,
		push.Payload.Repository.Slug,
		toCommitHash,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	query := req.URL.Query()
	query.Add("since", fromCommitHash)
	query.Add("withComments", "false")
	if start != -1 {
		query.Add("start", strconv.Itoa(start))
	}
	req.URL.RawQuery = query.Encode()
	req.SetBasicAuth(settings.StashUsername, settings.StashToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	defer resp.Body.Close()

	var body APIResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return 0, err
	}
	if !body.IsLastPage {
		nextStart = *body.NextPageStart
	}
	for _, diff := range body.Diffs {
		push.ChangedFiles = append(push.ChangedFiles, diff.Destination.Path)
	}

	return
}

// NewPush creates a new Push
func NewPush(payload WebhookPayload) (*Push, error) {
	p := &Push{
		Payload:      payload,
		ChangedFiles: make([]string, 0),
	}

	for _, change := range payload.Changes {
		var nextStart int
		var err error
		for start := -1; start != 0; start = nextStart {
			nextStart, err = getFilesChanged(p, change.FromHash, change.ToHash, start)
			if err != nil {
				return nil, err
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
	return p.Payload.Repository.Slug
}

// Org returns the name of the project.
func (p *Push) Org() string {
	return p.Payload.Repository.Project.Key
}

// IsMaster detects if the branch is master.
func (p *Push) IsMaster() bool {
	for _, change := range p.Payload.Changes {
		if change.IsMaster() {
			return true
		}
	}
	return false
}
