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

package util

import (
	"fmt"

	"github.com/minio/minio/pkg/wildcard"
)

type GithubRateLimitErr struct {
	RateLimit int
	RateReset string
}

func (e *GithubRateLimitErr) Error() string {
	return fmt.Sprintf("Exceeded Githib rate limit: limit %d, reset %s", e.RateLimit, e.RateReset)
}

type GitHubFileNotFoundErr struct {
	Org  string
	Repo string
	Path string
}

func IsGitHubFileNotFoundErr(errString string) bool {
	pattern := "No file named * found in *"
	return wildcard.Match(pattern, errString)
}

func (e *GitHubFileNotFoundErr) Error() string {
	return fmt.Sprintf("File %s not found for org %s in repository %s", e.Path, e.Org, e.Repo)
}
