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

type GitHubTest struct {
	contents string
	endpoint string
	err      error
}

func (g *GitHubTest) DownloadContents(org, repo, path, branch string) (string, error) {
	return g.contents, g.err
}

func (g *GitHubTest) CreateStatus(status *Status, org, repo, ref string) error {
	return g.err
}

func (g *GitHubTest) GetEndpoint() string {
	return g.endpoint
}
