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
	"errors"
	"github.com/armory/dinghy/pkg/git"
)

// SetCommitStatus sets the commit status
// TODO: this function needs to return an error but it's currently attached to an interface that does not
// and changes will affect other types
func (p *Push) SetCommitStatus(status git.Status, description string) {
	// no-op for GitLab
}

func (p *Push) GetCommitStatus() (error, git.Status, string) {
	return errors.New("functionality not implemented"), "",""
}