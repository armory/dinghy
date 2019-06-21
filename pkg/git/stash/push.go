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

package stash

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/armory/dinghy/pkg/git"
	"github.com/sirupsen/logrus"
)

// Push contains data about a push full of commits
type Push struct {
	Payload       WebhookPayload
	ChangedFiles  []string
	StashEndpoint string
	StashUsername string
	StashToken    string
	Logger        logrus.FieldLogger
}

// WebhookPayload is the payload from the webhook
type WebhookPayload struct {
	EventKey   string `json:"eventKey,omitempty"`
	Repository struct {
		Slug    string `json:"slug"`
		Project struct {
			Key string `json:"key"`
		} `json:"project"`
	} `json:"repository"`

	// bitbucket and old stash use different keys for `changes` but
	// the structure is the same
	BBSChanges   []WebhookChange `json:"changes"`
	StashChanges []WebhookChange `json:"refChanges"`
	IsOldStash   bool
}

// WebhookChange is a change object from the webhook
type WebhookChange struct {
	RefID    string `json:"refId"`
	FromHash string `json:"fromHash"`
	ToHash   string `json:"toHash"`
}

// IsMaster detects if a change was on master
func (c *WebhookChange) IsMaster() bool {
	return c.RefID == "refs/heads/master"
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

func (p *Push) getFilesChanged(fromCommitHash, toCommitHash string, start int) (nextStart int, err error) {
	url := fmt.Sprintf(
		`%s/projects/%s/repos/%s/commits/%s/changes`,
		p.StashEndpoint,
		p.Payload.Repository.Project.Key,
		p.Payload.Repository.Slug,
		toCommitHash,
	)
	p.Logger.Debug("ApiCall: ", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		p.Logger.Warnf("Could not create request for getFilesChanged (url: %s): %s", url, err.Error())
		return 0, err
	}

	query := req.URL.Query()
	query.Add("since", fromCommitHash)
	query.Add("withComments", "false")
	if start != -1 {
		query.Add("start", strconv.Itoa(start))
	}
	req.URL.RawQuery = query.Encode()
	req.SetBasicAuth(p.StashUsername, p.StashToken)

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		p.Logger.Warnf("getFilesChanged: failed to retrieve %s: %s", req.URL.RawQuery, err.Error())
		return 0, err
	}

	var body APIResponse
	p.Logger.Debugf("APIResponse: %+v\n", body)
	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("Got %d from retrieving commit data", resp.StatusCode)
		p.Logger.Error(msg)
		return 0, errors.New(msg)
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		p.Logger.Errorf("Got error parsing JSON response from Stash query %ss", url)
		return 0, err
	}
	if !body.IsLastPage {
		nextStart = body.NextPageStart
	}
	for _, diff := range body.Diffs {
		p.ChangedFiles = append(p.ChangedFiles, diff.Destination.Path)
	}

	return
}

type StashConfig struct {
	Username string
	Token    string
	Endpoint string
	Logger   logrus.FieldLogger
}

// NewPush creates a new Push
func NewPush(payload WebhookPayload, cfg StashConfig) (*Push, error) {
	p := &Push{
		Payload:       payload,
		ChangedFiles:  make([]string, 0),
		StashEndpoint: cfg.Endpoint,
		StashToken:    cfg.Token,
		StashUsername: cfg.Username,
		Logger:        cfg.Logger,
	}

	for _, change := range p.changes() {
		if !change.IsMaster() {
			continue
		}
		for start := -1; start != 0; {
			nextStart, err := p.getFilesChanged(change.FromHash, change.ToHash, start)
			if err != nil {
				cfg.Logger.Warnf("getFilesChanged failed: %s", err.Error())
				return nil, err
			}
			start = nextStart
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

func (p *Push) changes() []WebhookChange {
	if p.Payload.IsOldStash {
		return p.Payload.StashChanges
	}
	return p.Payload.BBSChanges
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
