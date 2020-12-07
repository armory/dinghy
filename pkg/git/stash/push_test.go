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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsFile(t *testing.T) {
	push := Push{ChangedFiles: []string{"Dinghyfile", "foo/bar/SubdirDinghyfile"}}

	assert.True(t, push.ContainsFile("Dinghyfile"))
	assert.True(t, push.ContainsFile("SubdirDinghyfile"))
	assert.False(t, push.ContainsFile("dinghyfile"))
}

func TestIsBranch(t *testing.T) {
	testCases := map[string]struct {
		webhookBranchName string
		configBranchName  string
		expected          bool
	}{
		"true": {
			webhookBranchName: "refs/heads/some_branch",
			configBranchName:  "some_branch",
			expected:          true,
		},
		"true again": {
			webhookBranchName: "refs/heads/some_branch",
			configBranchName:  "refs/heads/some_branch",
			expected:          true,
		},
		"false": {
			webhookBranchName: "refs/heads/some_branch",
			configBranchName:  "meh",
			expected:          false,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			payload := fmt.Sprintf(`{"changes": [{"refId": "%s"}]}`, tc.webhookBranchName)
			webhookPayload := WebhookPayload{}
			if err := json.NewDecoder(bytes.NewBufferString(payload)).Decode(&webhookPayload); err != nil {
				t.Fatalf(err.Error())
			}

			p := &Push{Payload: webhookPayload}

			actual := p.IsBranch(tc.configBranchName)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
