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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrg(t *testing.T) {
	t.Skip()
	cases := []struct {
		payload  string
		expected string
	}{
		{
			payload:  `{"repository": {"organization": "org-armory"}}`,
			expected: "org-armory",
		},
		{
			payload:  `{"repository": {"owner": {"login": "login-armory"}}}`,
			expected: "login-armory",
		},
		{
			payload:  `{"repository": { "organization": "org-armory", "owner": {"login": "login-armory"}}}}`,
			expected: "org-armory",
		},
		{
			payload:  `{"EventKey": ""}`,
			expected: "",
		},
	}

	for _, c := range cases {
		var p Push
		if err := json.NewDecoder(bytes.NewBufferString(c.payload)).Decode(&p); err != nil {
			t.Fatalf(err.Error())
		}

		if p.Org() != c.expected {
			t.Fatalf("failed to verify that %s matches %s", p.Org(), c.expected)
		}
	}
}

func TestDefaultBranch(t *testing.T) {
	p := Push{
		Payload: WebhookPayload{
			IsOldStash: true,
			StashChanges: []WebhookChange{
				WebhookChange{RefID: "refs/heads/default"},
				WebhookChange{RefID: "refs/heads/other"},
			},
		},
	}
	assert.Equal(t, p.IsMaster("default"), true)
	assert.Equal(t, p.IsMaster("other"), true)
	assert.Equal(t, p.IsMaster(""), false)
	assert.Equal(t, p.IsMaster("master"), false)

	p.Payload.StashChanges = append(p.Payload.StashChanges, WebhookChange{RefID: "refs/heads/master"})
	assert.Equal(t, p.IsMaster(""), true)
	assert.Equal(t, p.IsMaster("master"), true)
	assert.Equal(t, p.IsMaster("nothing"), false)
}
