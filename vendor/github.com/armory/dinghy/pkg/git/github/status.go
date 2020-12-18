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

package github

import (
	"github.com/armory/dinghy/pkg/git"
)

/* Example: POST /repos/:owner/:repo/statuses/:sha
{
  "state": "success",
  "target_url": "https://example.com/build/status",
  "description": "The build succeeded!",
  "context": "continuous-integration/jenkins"
} */

// Status is a type that contains information relevant to the commit status
type Status struct {
	State       string
	DeckBaseURL string
	Description string
	Context     string
}

// SetCommitStatus sets the commit status
// TODO: this function needs to return an error but it's currently attached to an interface that does not
// and changes will affect other types
func (p *Push) SetCommitStatus(status git.Status, description string) {
	for _, c := range p.Commits {
		s := newStatus(status, p.DeckBaseURL, description)
		if err := p.Config.CreateStatus(s, p.Org(), p.Repo(), c.ID); err != nil {
			p.Logger.Error(err)
			return
		}
	}
}

func (p *Push) GetCommitStatus() (error, git.Status, string) {
	err, statuses := p.Config.ListStatuses(p.Org(), p.Repo(), p.Branch())
	if err != nil {
		p.Logger.Warnf("Failed to get status information for %v/%v/%v", p.Org(), p.Repo(), p.Branch())
		return err, "", ""
	}
	for _, status := range statuses {
		if status.GetContext() == "dinghy" {
			return nil, git.Status(status.GetState()), status.GetDescription()
		}
	}
	return nil, "", ""
}

// Commits return the list of commit hashes
func (p *Push) GetCommits() []string {
	var result []string
	for _, val := range p.Commits {
		result = append(result, val.ID)
	}
	return result
}

func newStatus(s git.Status, deckURL string, description string) *Status {
	state := string(s)
	context := "dinghy"

	if len(description) > 140 {
		description = description[0:136] + "..."
	}

	return &Status{
		State:       state,
		DeckBaseURL: deckURL,
		Context:     context,
		Description: description,
	}
}
