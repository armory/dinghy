package github

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func newGitHubClient(ctx context.Context, endpoint, token string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client, err := github.NewEnterpriseClient(endpoint, endpoint, tc)
	if err != nil {
		return nil, err
	}

	return client, nil
}
