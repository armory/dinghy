package github

import (
	"fmt"
	"github.com/armory-io/dinghy/pkg/git"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCommitStatusSuccessfully(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "")
	}))
	defer ts.Close()
	// TODO: Do not use global variable. This will lead to side-effects.
	p := Push{
		GitHubEndpoint: ts.URL,
		Repository:     Repository{Organization: "armory-io", Name: "dinghy"},
		Commits: []Commit{
			{
				ID: "ABC",
			},
		},
	}

	// This shouldn't throw exceptions/panics
	p.SetCommitStatus(git.StatusPending)
}

func TestSetCommitStatusFails(t *testing.T) {
	// TODO: Do not use global variable. This will lead to side-effects.
	p := Push{
		GitHubEndpoint: "invalid-url",
		Repository:     Repository{Organization: "armory-io", Name: "dinghy"},
		Commits: []Commit{
			{
				ID: "ABC",
			},
		},
	}

	// This shouldn't throw exceptions/panics
	p.SetCommitStatus(git.StatusPending)
}
