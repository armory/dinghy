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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/armory/dinghy/pkg/git"
	"github.com/armory/dinghy/pkg/util"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

const (
	head_commit = "head_commit"
	id          = "id"
)

type GitHubClient interface {
	DownloadContents(string, string, string, string) (string, error)
	CreateStatus(*Status, string, string, string) error
	GetEndpoint() string
	GetToken() string
}

type Config struct {
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

func (g *Config) DownloadContents(org, repo, path, branch string) (string, error) {
	ctx := context.Background()
	client, err := newGitHubClient(ctx, g.Endpoint, g.Token)
	if err != nil {
		return "", err
	}

	opt := &github.RepositoryContentGetOptions{Ref: branch}
	r, _, err := client.Repositories.DownloadContents(ctx, org, repo, path, opt)
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

func (g *Config) CreateStatus(status *Status, org, repo, ref string) error {
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

func (g *Config) ListStatuses(org, repo, ref string) (error, []*github.RepoStatus) {

	ctx := context.Background()
	client, err := newGitHubClient(ctx, g.Endpoint, g.Token)
	if err != nil {
		return err, nil
	}

	repoStatuses, _, errorStatus := client.Repositories.ListStatuses(ctx, org, repo, ref, nil)
	if errorStatus != nil {
		if e, ok := errorStatus.(*github.RateLimitError); ok {
			return &util.GithubRateLimitErr{RateLimit: e.Rate.Limit, RateReset: e.Rate.Reset.String()}, nil
		}
		return errorStatus, nil
	}

	return nil, repoStatuses
}

func (g *Config) GetPullRequest(org, repo, ref, sha string) (*github.PullRequest, error) {

	ctx := context.Background()
	client, err := newGitHubClient(ctx, g.Endpoint, g.Token)
	if err != nil {
		return nil, err
	}

	pullRequests, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, org, repo, sha, &github.PullRequestListOptions{Base: ref})

	if err != nil {
		return nil, err
	}
	if len(pullRequests) > 0 {
		return pullRequests[0], nil
	} else {
		return nil, nil
	}
}

func (g *Config) GetShaFromRawData(rawPushData []byte) string {

	// deserialze push data to a map.  used in template logic later
	content := make(map[string]interface{})
	_ = json.Unmarshal(rawPushData, &content)

	sha := ""
	v, ok := content[head_commit]
	if ok {
		v, ok = v.(map[string]interface{})[id]
		if ok {
			sha = v.(string)
		}
	}

	return sha
}

func (g *Config) GetEndpoint() string {
	return g.Endpoint
}

func (g *Config) GetToken() string {
	return g.Token
}

func (g *Config) Slug(pushEvent map[string]interface{}) (*git.RepositoryInfo, error) {

	if htmlURL, found := pushEvent["html_url"]; found {
		if repo, ok := htmlURL.(string); ok {
			repoUrl, err := url.Parse(repo)
			if err != nil {
				return nil, errors.New("push event does not contain a valid html url")
			}
			splitRepo := strings.Split(repoUrl.Path, "/")
			if len(splitRepo) < 2 {
				// Implies that our HTML URL does not parse cleanly into a git project/url string.
				return nil, errors.New("unable to parse repository, not enough segments")
			}

			return &git.RepositoryInfo{Org: splitRepo[1], Repo: splitRepo[2], Type: "github"}, nil
		}
	}

	return nil, errors.New("unable to find html url from push event")
}
