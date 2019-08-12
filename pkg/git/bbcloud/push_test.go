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

package bbcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const webhookPayloadOneChange = `{
  "push": {
    "changes": [
      {
        "old": {
          "name": "%s",
          "target": {
            "hash": "4057376176a3f4fcaf4e52b11b57c76ac212542e"
          }
        },
        "new": {
          "name": "%s",
          "target": {
            "hash": "4d2dc6d19baf423c8b94c281a57f932992fdb1f6"
          }
        }
      }
    ]
  },
  "repository": {
    "name": "myRepo",
    "full_name": "myOrg/myRepo",
	"project": {
		"key": "myProject"
	}
  }
}`

const webhookPayloadTwoChanges = `{
  "push": {
    "changes": [
      {
        "old": {
          "name": "master",
          "target": {
            "hash": "4057376176a3f4fcaf4e52b11b57c76ac212542d"
          }
        },
        "new": {
          "name": "master",
          "target": {
            "hash": "4d2dc6d19baf423c8b94c281a57f932992fdb1f5"
          }
        }
      },
      {
        "old": {
          "name": "feature/awesome",
          "target": {
            "hash": "4057376176a3f4fcaf4e52b11b57c76ac212542e"
          }
        },
        "new": {
          "name": "master",
          "target": {
            "hash": "4d2dc6d19baf423c8b94c281a57f932992fdb1f6"
          }
        }
      }
    ]
  },
  "repository": {
    "name": "myRepo",
    "full_name": "myOrg/myRepo",
	"project": {
		"key": "myProject"
	}
  }
}`

const diffstatResponseOneFile = `{
  "pagelen": 1,
  "values": [
    {
      "old": {
        "path": "%s"
      },
      "new": {
        "path": "%s"
      }
    }
  ],
  "page": %d,
  "size":%d 
}`

const diffstatResponseTwoFiles = `{
  "pagelen": 2,
  "values": [
    {
      "old": {
        "path": "dinghyfile"
      },
      "new": {
        "path": "dinghyfile"
      }
    },
    {
      "old": {
        "path": "Jenkinsfile"
      },
      "new": {
        "path": "Jenkinsfile"
      }
    }
  ],
  "page": 1,
  "size": 1
}`

func TestNewPushIncludesDinghyfileRenamed(t *testing.T) {
	webhookPayload := WebhookPayload{}
	payloadString := fmt.Sprintf(webhookPayloadOneChange, "master", "master")
	if err := json.NewDecoder(bytes.NewBufferString(payloadString)).Decode(&webhookPayload); err != nil {
		t.Fatalf(err.Error())
	}
	diffStatResponse := fmt.Sprintf(diffstatResponseOneFile, "dinghyfile_bkp", "dinghyfile", 1, 1)

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if _, err := res.Write([]byte(diffStatResponse)); err != nil {
			t.Fatalf(err.Error())
		}
	}))
	defer func() { testServer.Close() }()

	push, err := NewPush(webhookPayload, Config{Endpoint: testServer.URL})
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 1, len(push.ChangedFiles))
	assert.True(t, contains(push.ChangedFiles, "dinghyfile"), "Error: expected dinghyfile found in push info")
}

func TestNewPushTwoCommitsToSameFile(t *testing.T) {
	webhookPayload := WebhookPayload{}
	if err := json.NewDecoder(bytes.NewBufferString(webhookPayloadTwoChanges)).Decode(&webhookPayload); err != nil {
		t.Fatalf(err.Error())
	}
	diffStatResponse := fmt.Sprintf(diffstatResponseOneFile, "dinghyfile", "dinghyfile", 1, 1)

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if _, err := res.Write([]byte(diffStatResponse)); err != nil {
			t.Fatalf(err.Error())
		}
	}))
	defer func() { testServer.Close() }()

	push, err := NewPush(webhookPayload, Config{Endpoint: testServer.URL})
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 1, len(push.ChangedFiles))
	assert.True(t, contains(push.ChangedFiles, "dinghyfile"), "Error: expected dinghyfile found in push info")
}

func TestNewPushIncludesMultipleChangedFiles(t *testing.T) {
	webhookPayload := WebhookPayload{}
	payloadString := fmt.Sprintf(webhookPayloadOneChange, "master", "master")
	if err := json.NewDecoder(bytes.NewBufferString(payloadString)).Decode(&webhookPayload); err != nil {
		t.Fatalf(err.Error())
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if _, err := res.Write([]byte(diffstatResponseTwoFiles)); err != nil {
			t.Fatalf(err.Error())
		}
	}))
	defer func() { testServer.Close() }()

	push, err := NewPush(webhookPayload, Config{Endpoint: testServer.URL})
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 2, len(push.ChangedFiles))
	assert.True(t, contains(push.ChangedFiles, "dinghyfile"), "Error: expected dinghyfile found in push info")
	assert.True(t, contains(push.ChangedFiles, "Jenkinsfile"), "Error: expected Jenkinsfile found in push info")
}

func TestNewPushWithPagination(t *testing.T) {
	webhookPayload := WebhookPayload{}
	payloadString := fmt.Sprintf(webhookPayloadOneChange, "master", "master")
	if err := json.NewDecoder(bytes.NewBufferString(payloadString)).Decode(&webhookPayload); err != nil {
		t.Fatalf(err.Error())
	}
	diffStatResponses := []string{
		fmt.Sprintf(diffstatResponseOneFile, "dinghyfile", "dinghyfile", 1, 2),
		fmt.Sprintf(diffstatResponseOneFile, "Main.java", "Main.java", 2, 2),
	}
	diffStatResponseIndex := 0

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if _, err := res.Write([]byte(diffStatResponses[diffStatResponseIndex])); err != nil {
			t.Fatalf(err.Error())
		}
		diffStatResponseIndex++
	}))
	defer func() { testServer.Close() }()

	push, err := NewPush(webhookPayload, Config{Endpoint: testServer.URL})
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 2, len(push.ChangedFiles))
	assert.True(t, contains(push.ChangedFiles, "dinghyfile"), "Error: expected dinghyfile found in push info")
	assert.True(t, contains(push.ChangedFiles, "Main.java"), "Error: expected Main.java found in push info")
}

func TestNewPushFromFeatureBranchToMaster(t *testing.T) {
	webhookPayload := WebhookPayload{}
	payloadString := fmt.Sprintf(webhookPayloadOneChange, "feature/awesome", "master")
	if err := json.NewDecoder(bytes.NewBufferString(payloadString)).Decode(&webhookPayload); err != nil {
		t.Fatalf(err.Error())
	}
	diffStatResponse := fmt.Sprintf(diffstatResponseOneFile, "dinghyfile_bkp", "dinghyfile", 1, 1)

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if _, err := res.Write([]byte(diffStatResponse)); err != nil {
			t.Fatalf(err.Error())
		}
	}))
	defer func() { testServer.Close() }()

	push, err := NewPush(webhookPayload, Config{Endpoint: testServer.URL})
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 1, len(push.ChangedFiles))
	assert.True(t, contains(push.ChangedFiles, "dinghyfile"), "Error: expected dinghyfile found in push info")
}

func contains(files []string, file string) bool {
	for _, a := range files {
		if a == file {
			return true
		}
	}
	return false
}

func TestIsBranch(t *testing.T) {
	testCases := map[string]struct {
		webhookBranchName         string
		configBranchName        string
		expected    bool
	}{
		"true": {
			webhookBranchName: "refs/heads/some_branch",
			configBranchName: "some_branch",
			expected: true,
		},
		"true again": {
			webhookBranchName: "refs/heads/some_branch",
			configBranchName: "refs/heads/some_branch",
			expected: true,
		},
		"false": {
			webhookBranchName: "refs/heads/some_branch",
			configBranchName: "meh",
			expected: false,
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			webhookPayload := WebhookPayload{}
			payloadString := fmt.Sprintf(webhookPayloadOneChange, tc.webhookBranchName, tc.webhookBranchName)
			if err := json.NewDecoder(bytes.NewBufferString(payloadString)).Decode(&webhookPayload); err != nil {
				t.Fatalf(err.Error())
			}
			diffStatResponse := fmt.Sprintf(diffstatResponseOneFile, "dinghyfile_bkp", "dinghyfile", 1, 1)

			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				if _, err := res.Write([]byte(diffStatResponse)); err != nil {
					t.Fatalf(err.Error())
				}
			}))
			defer func() { testServer.Close() }()

			push, err := NewPush(webhookPayload, Config{Endpoint: testServer.URL})
			if err != nil {
				t.Fatalf(err.Error())
			}

			actual := push.IsBranch(tc.configBranchName)
			assert.Equal(t, tc.expected, actual)
		})
	}
}