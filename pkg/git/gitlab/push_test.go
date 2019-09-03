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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func loadExample(t *testing.T) string {
	content, err := ioutil.ReadFile("example_payload.json")
	require.Equal(t, err, nil)
	return string(content)
}

func TestOrg(t *testing.T) {
	cases := []struct {
		payload  string
		expected string
	}{
		{
			payload:  loadExample(t),
			expected: "mike",
		},
	}

	for _, c := range cases {
		var p Push
		err := json.NewDecoder(bytes.NewBufferString(c.payload)).Decode(&p)
		require.Nil(t, err)
		assert.Equal(t, p.Org(), c.expected)
	}
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
			payload := fmt.Sprintf(`{"repository": {"organization": "org-armory"}, "ref": "%s"}`, tc.webhookBranchName)
			var p Push
			err := json.NewDecoder(bytes.NewBufferString(payload)).Decode(&p)
			assert.Nil(t, err)

			actual := p.IsBranch(tc.configBranchName)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
