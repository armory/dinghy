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

package dummy

import "github.com/armory/dinghy/pkg/git"

// Push contains data about a push full of commits
type Push struct {
	RepoName  string
	OrgName   string
	FileNames []string
}

// ContainsFile checks to see if a given file is in the push.
func (p *Push) ContainsFile(file string) bool {
	for _, name := range p.FileNames {
		if name == file {
			return true
		}
	}
	return false
}

// Files returns a slice containing filenames that were added/modified
func (p *Push) Files() []string {
	return p.FileNames
}

// Repo returns the name of the repo.
func (p *Push) Repo() string {
	return p.RepoName
}

// Org returns the name of the project.
func (p *Push) Org() string {
	return p.OrgName
}

// Branch returns the branch of the push
func (p *Push) Branch() string {
	return "branch_name"
}

// SetCommitStatus sets a commit status
func (p *Push) SetCommitStatus(s git.Status) {
	// Do nothing
}

// Name returns the name of the provider to be used in configuration
func (p *Push) Name() string {
	return "dummy"
}
