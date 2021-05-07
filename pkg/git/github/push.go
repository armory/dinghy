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
	"strings"

	"github.com/armory/dinghy/pkg/log"
)

// Push is the payload received from a GitHub webhook.
type Push struct {
	Commits     []Commit   `json:"commits"`
	Repository  Repository `json:"repository"`
	Ref         string     `json:"ref"`
	Config      Config
	DeckBaseURL string
	Logger      log.DinghyLog
}

// Commit is a commit received from Github webhook
type Commit struct {
	ID       string   `json:"id"`
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
}

// Repository is a repo received from Github webhook
type Repository struct {
	Name         string          `json:"name"`
	Organization string          `json:"organization"`
	Owner        RepositoryOwner `json:"owner"`
	FullName     string          `json:"full_name"`
}

type RepositoryOwner struct {
	Login string `json:"login"`
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
	if p.Commits == nil {
		return false
	}
	for _, c := range p.Commits {
		if inSlice(c.Added, file) || inSlice(c.Modified, file) {
			return true
		}
	}
	return false
}

// Files returns a slice containing filenames that were added/modified
func (p *Push) Files() []string {
	ret := make([]string, 0, 0)
	if p.Commits == nil {
		return ret
	}
	for _, c := range p.Commits {
		ret = append(ret, c.Added...)
		ret = append(ret, c.Modified...)
	}
	return ret
}

// Repo returns the name of the repo.
func (p *Push) Repo() string {
	return p.Repository.Name
}

// Org returns the organization of the push
func (p *Push) Org() string {
	segments := strings.Split(p.Repository.FullName, "/")
	return strings.Join(segments[:len(segments)-1], "/")
}

// Branch returns the branch of the push
func (p *Push) Branch() string {
	return p.Ref
}

func (p *Push) IsBranch(branchToTry string) bool {
	// our configuration only requires the branch name, but the branch comes
	// from the webhook as "refs/heads/branch"
	return strings.Replace(p.Branch(), "refs/heads/", "", 1) == strings.Replace(branchToTry, "refs/heads/", "", 1)
}

// IsMaster detects if the branch is master.
// Main is the new master
func (p *Push) IsMaster() bool {
	return p.Ref == "master" || p.Ref == "refs/heads/master" || p.Ref == "main" || p.Ref == "refs/heads/main"
}

// Name returns the name of the provider to be used in configuration
func (p *Push) Name() string {
	return "github"
}
