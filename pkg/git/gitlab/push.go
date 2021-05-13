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

package gitlab

import (
	"github.com/armory/dinghy/pkg/log"
	"github.com/armory/dinghy/pkg/settings/global"
	gitlab "github.com/xanzy/go-gitlab"
	"strings"
)

// Push is the payload received from a GitHub webhook.
type Push struct {
	Event       *gitlab.PushEvent
	DeckBaseURL string
	Logger      log.DinghyLog
}

func inSlice(arr []string, val string) bool {
	for _, filePath := range arr {
		// eg: filePath can be: "app/src/dinghyfile"
		//     or just "dinghyfile"
		components := strings.Split(filePath, "/")
		if components[len(components)-1] == val {
			return true
		}
	}
	return false
}

// ContainsFile checks to see if a given file is in the push.
func (p *Push) ContainsFile(file string) bool {
	if p.Event.Commits == nil {
		return false
	}
	for _, c := range p.Event.Commits {
		if inSlice(c.Added, file) || inSlice(c.Modified, file) {
			return true
		}
	}
	return false
}

// Files returns a slice containing filenames that were added/modified
func (p *Push) Files() []string {
	ret := make([]string, 0, 0)
	if p.Event.Commits == nil {
		return ret
	}
	for _, c := range p.Event.Commits {
		ret = append(ret, c.Added...)
		ret = append(ret, c.Modified...)
	}
	return ret
}

// Repo returns the name of the repo.
func (p *Push) Repo() string {
	return p.Event.Project.Name
}

// Org returns the organization of the push
func (p *Push) Org() string {
	return strings.SplitN(p.Event.Project.PathWithNamespace, "/", 2)[0]
}

// Branch returns the branch of the push
func (p *Push) Branch() string {
	return p.Event.Ref
}

func (p *Push) IsBranch(branchToTry string) bool {
	// our configuration only requires the branch name, but the branch comes
	// from the webhook as "refs/heads/branch"
	return strings.Replace(p.Branch(), "refs/heads/", "", 1) == strings.Replace(branchToTry, "refs/heads/", "", 1)
}

// IsMaster detects if the branch is master.
func (p *Push) IsMaster(dinghyfileBranches []string) bool {
	for _, branch := range dinghyfileBranches {
		if p.Event.Ref == "refs/heads/"+branch {
			return true
		}
	}
	return p.Event.Ref == "refs/heads/master" || p.Event.Ref == "refs/heads/main"
}

// Name returns the name of the provider to be used in configuration
func (p *Push) Name() string {
	return "gitlab"
}

// ParseWebhook parses the webhook into the struct and returns a file service
// instance (and error)
func (p *Push) ParseWebhook(cfg *global.Settings, body []byte) (FileService, error) {
	fs := FileService{
		Logger: p.Logger,
		Client: gitlab.NewClient(nil, cfg.GitLabToken),
	}

	// Note:  SetBaseURL will ensure a trailing slash as needed.
	err := fs.Client.SetBaseURL(cfg.GitLabEndpoint)
	if err != nil {
		return fs, err
	}

	// Let go-gitlab do all the work.
	event, err := gitlab.ParseWebhook(gitlab.EventTypePush, body)
	if err != nil {
		return fs, err
	}
	p.Event = event.(*gitlab.PushEvent)
	return fs, nil
}
