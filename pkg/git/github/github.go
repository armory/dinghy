package github

import (
	"bytes"
	"context"
	"fmt"

	"github.com/armory-io/dinghy/pkg/util"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GitHubClient interface {
	DownloadContents(string, string, string) (string, error)
	CreateStatus(*Status, string, string, string) error
	GetEndpoint() string
}

type GitHub struct {
	Endpoint string
	Token    string
}

func newGitHubClient(ctx context.Context, endpoint, token string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client, err := github.NewEnterpriseClient(endpoint, endpoint, tc)
	if err != nil {
		return nil, fmt.Errorf("unable to create github client: %s", err)
	}

	return client, nil
}

func (g *GitHub) DownloadContents(org, repo, path string) (string, error) {
	ctx := context.Background()
	client, err := newGitHubClient(ctx, g.Endpoint, g.Token)
	if err != nil {
		return "", err
	}

	r, err := client.Repositories.DownloadContents(ctx, org, repo, path, nil)
	if err != nil {
		if e, ok := err.(*github.RateLimitError); ok {
			return "", &util.GithubRateLimitErr{RateLimit: e.Rate.Limit, RateReset: e.Rate.Reset.String()}
		}
		if util.IsGitHubFileNotFoundErr(err.Error()) {
			return "", &util.GitHubFileNotFoundErr{Org: org, Repo: repo, Path: path}
		}
		return "", err
	}

	b := new(bytes.Buffer)
	b.ReadFrom(r)
	r.Close()

	return b.String(), nil
}

func (g *GitHub) CreateStatus(status *Status, org, repo, ref string) error {
	repoStatus := &github.RepoStatus{
		State:       &status.State,
		TargetURL:   &status.DeckBaseURL,
		Context:     &status.Context,
		Description: &status.Description,
	}

	ctx := context.Background()
	client, err := newGitHubClient(ctx, g.Endpoint, g.Token)
	if err != nil {
		return err
	}

	_, _, err = client.Repositories.CreateStatus(ctx, org, repo, ref, repoStatus)
	if err != nil {
		if e, ok := err.(*github.RateLimitError); ok {
			return &util.GithubRateLimitErr{RateLimit: e.Rate.Limit, RateReset: e.Rate.Reset.String()}
		}
		return err
	}
	return nil
}

func (g *GitHub) GetEndpoint() string {
	return g.Endpoint
}
