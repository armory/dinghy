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

package git

import pipebuilder "github.com/armory/dinghy/pkg/dinghyfile/pipebuilder"

// Status wires up to the green check or red x next to a GitHub commit.
type Status string

// Status types
const (
	StatusPending                 Status = "pending"
	StatusError                          = "error"
	StatusSuccess                        = "success"
	StatusFailure                        = "failure"
	DefaultPendingMessage         string = "Updating pipeline definitions..."
	DefaultSuccessMessage         string = "Pipeline definitions updated!"
	DefaultValidatePendingMessage string = "Validating pipeline definitions..."
	DefaultValidateSuccessMessage string = "Pipeline definitions validation was successful!"
)

var DefaultMessagesByBuilderAction = map[pipebuilder.BuilderAction]map[Status]string{
	pipebuilder.Process: {
		StatusPending: DefaultPendingMessage,
		StatusSuccess: DefaultSuccessMessage,
	},
	pipebuilder.Validate: {
		StatusPending: DefaultValidatePendingMessage,
		StatusSuccess: DefaultValidateSuccessMessage,
	},
}
