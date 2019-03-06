package bbcloud

import (
	"bytes"
	"encoding/json"
	"github.com/magiconair/properties/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const genericWebhookPayload = `{
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

const genericDiffstatResponse = `{
  "pagelen": 500,
  "values": [
    {
      "old": {
        "path": "dinghyfile_bkp"
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

func TestNewPushIncludesDinghyfile(t *testing.T) {
	webhookPayload := WebhookPayload{}
	if err := json.NewDecoder(bytes.NewBufferString(genericWebhookPayload)).Decode(&webhookPayload); err != nil {
		t.Fatalf(err.Error())
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if _, err := res.Write([]byte(genericDiffstatResponse)); err != nil {
			t.Fatalf(err.Error())
		}
	}))
	defer func() { testServer.Close() }()

	push, err := NewPush(webhookPayload, Config{Endpoint: testServer.URL})
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Equal(t, 2, len(push.ChangedFiles))
	assert.Equal(t, push.ChangedFiles[0], "dinghyfile", "Error: expected dinghyfile found in push info")
	assert.Equal(t, push.ChangedFiles[1], "Jenkinsfile", "Error: expected Jenkinsfile found in push info")
}
