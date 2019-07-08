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

package github

import (
	"bytes"
	"context"
	"fmt"
	"github.com/armory/dinghy/pkg/util"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GitHubClient interface {
	DownloadContents(string, string, string, string) (string, error)
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

func (g *GitHub) DownloadContents(org, repo, path, branch string) (string, error) {
	ctx := context.Background()
	client, err := newGitHubClient(ctx, g.Endpoint, g.Token)
	if err != nil {
		return "", err
	}

	opt := github.RepositoryContentGetOptions{Ref: branch}
	r, err := client.Repositories.DownloadContents(ctx, org, repo, path, &opt)
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
