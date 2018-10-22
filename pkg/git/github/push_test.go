package github

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestOrg(t *testing.T) {
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
