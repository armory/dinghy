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
	"github.com/sirupsen/logrus"
	"strings"
)

// Push is the payload received from a GitHub webhook.
type Push struct {
	Commits     []Commit   `json:"commits"`
	Repository  Repository `json:"repository"`
	Ref         string     `json:"ref"`
	GitHub      GitHub
	DeckBaseURL string
	Logger      logrus.FieldLogger
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
	if p.Repository.Organization != "" {
		return p.Repository.Organization
	}

	return p.Repository.Owner.Login
}

// IsMaster detects if the branch is master.
func (p *Push) IsMaster() bool {
	return p.Ref == "refs/heads/master"
}
