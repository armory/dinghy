package github

import (
	"context"

	"github.com/google/go-github/github"

	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/util"

	log "github.com/sirupsen/logrus"
)

/* Example: POST /repos/:owner/:repo/statuses/:sha
{
  "state": "success",
  "target_url": "https://example.com/build/status",
  "description": "The build succeeded!",
  "context": "continuous-integration/jenkins"
} */

// Status is a payload that we send to the Github API to set commit status
type Status struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// SetCommitStatus sets the commit status
// TODO: this function needs to return an error but it's currently attached to an interface that does not
// and changes will affect other types
func (p *Push) SetCommitStatus(status git.Status) {
	ctx := context.Background()
	client, err := newGitHubClient(ctx, p.GitHubEndpoint, p.GitHubToken)
	if err != nil {
		log.Errorf("unable to create github client for push SetCommitStatus request: %s", err)
	}

	repoStatus := newStatus(status, p.DeckBaseURL)
	for _, c := range p.Commits {
		_, _, err := client.Repositories.CreateStatus(ctx, p.Org(), p.Repo(), c.ID, repoStatus)
		if err != nil {
			if e, ok := err.(*github.RateLimitError); ok {
				rlErr := util.GithubRateLimitErr{RateLimit: e.Rate.Limit, RateReset: e.Rate.Reset.String()}
				log.Errorf(rlErr.Error())
				return
			}
			log.Error(err)
			return
			//return err
		}
	}

	//return nil
}

func newStatus(s git.Status, deckURL string) *github.RepoStatus {
	state := string(s)
	context := "continuous-deployment/dinghy"
	description := ""
	switch s {
	case git.StatusSuccess:
		description = "Pipeline definitions updated!"
	case git.StatusError:
		description = "Error updating pipeline definitions!"
	case git.StatusFailure:
		description = "Failed to update pipeline definitions!"
	case git.StatusPending:
		description = "Updating pipeline definitions..."
	}

	return &github.RepoStatus{
		State:       &state,
		TargetURL:   &deckURL,
		Context:     &context,
		Description: &description,
	}
}
