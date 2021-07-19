package notifiers

import (
	"bytes"
	ctx "context"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory/plank/v4"
	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	logevent    = "logevent"
	rawdata     = "rawdata"
	head_commit = "head_commit"
	id          = "id"
)

type GithubNotificationValues struct {
	When    []string `json:"when" yaml:"when" mapstructure:"when"`
	Address string   `json:"address" yaml:"address" mapstructure:"address"`
	Level   string   `json:"level" yaml:"level" mapstructure:"level"`
}

type GithubNotifier struct {
	client *github.Client
}

func (s *GithubNotificationValues) GetWhen() []string {
	return s.When
}

func (s *GithubNotificationValues) GetLevel() string {
	return s.Level
}

func (s *GithubNotificationValues) GetAddress() string {
	return s.Address
}

type GithubNotification struct {
	Body  string
	Owner string
	Repo  string
	Sha   string
}

func generateNotifications(owner, repo string, content map[string]interface{}, isError bool, message string) []GithubNotification {
	charLimit := 65000
	body := ""
	if v, ok := content[logevent]; ok {
		body = v.(string)
	}

	sha := ""
	if v, ok := content[rawdata]; ok {
		v, ok = v.(map[string]interface{})[head_commit]
		if ok {
			v, ok = v.(map[string]interface{})[id]
			if ok {
				sha = v.(string)
			}
		}
	}
	numCommentsNeeded := int(math.Ceil(float64(len(body)) / float64(charLimit)))
	githubNotifications := []GithubNotification{}
	var commentBody string
	for i := 0; i < numCommentsNeeded; i++ {
		if i == 0 {
			if numCommentsNeeded == 1 {
				commentBody = buildFirstComment(body, message, isError)
			} else {
				commentBody = buildFirstComment(body[0:charLimit], message, isError)
			}
		} else {
			startingLoc := (i * charLimit)
			if i == numCommentsNeeded-1 {
				commentBody = buildSubsequentComments(body[startingLoc:])
			} else {
				commentBody = buildSubsequentComments(body[startingLoc : startingLoc+charLimit])
			}
		}

		githubNotifications = append(githubNotifications, GithubNotification{
			Body:  commentBody,
			Owner: owner,
			Repo:  repo,
			Sha:   sha,
		})
	}
	return githubNotifications
}

func NewGithubNotifier(s *settings.ExtSettings) (*GithubNotifier, error) {
	ghClient, err := newGithubClient(s.Settings.GitHubToken, s.Settings.GithubEndpoint)

	if err != nil {
		return nil, err
	}

	return &GithubNotifier{
		client: ghClient,
	}, nil
}

func (gn *GithubNotifier) SendOnValidation() bool {
	return true
}

func (gn *GithubNotifier) SendSuccess(org, repo, path string, notificationsType plank.NotificationsType, content map[string]interface{}) {
	notifications := generateNotifications(org, repo, content, false, fmt.Sprintf("%s/%s/%s", org, repo, path))
	for _, notification := range notifications {
		err := gn.createCommitComment(notification.Body, notification.Owner, notification.Repo, notification.Sha, false)
		if err != nil {
			log.Error(err)
		}
	}
}

func (gn *GithubNotifier) SendFailure(org, repo, path string, errorDinghy error, notificationsType plank.NotificationsType, content map[string]interface{}) {
	notifications := generateNotifications(org, repo, content, true, fmt.Sprintf("%s/%s/%s (%s)", org, repo, path, errorDinghy.Error()))
	for _, notification := range notifications {
		err := gn.createCommitComment(notification.Body, notification.Owner, notification.Repo, notification.Sha, true)
		if err != nil {
			log.Error(err)
		}
	}
}

const (
	// PublicGithubEndpoint is the default endpoint for GitHub.
	PublicGithubEndpoint string = "https://api.github.com"
)

func newGithubClient(accessToken, baseURL string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx.Background(), ts)

	if baseURL == "" {
		baseURL = PublicGithubEndpoint
	}

	// Cribbed from: https://github.com/google/go-github/blob/b338ce623f50aeaf4cbf7f4195a79d0d302f9a65/github/github.go#L318-L324
	baseEndpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseEndpoint.Path, "/") {
		baseEndpoint.Path += "/"
	}

	ghClient := github.NewClient(tc)
	ghClient.BaseURL = baseEndpoint

	return ghClient, nil
}

func (gn *GithubNotifier) createCommitComment(body, owner, repo, sha string, isError bool) error {

	input := &github.RepositoryComment{
		Body: github.String(body),
	}

	comment, response, err := gn.client.Repositories.CreateComment(ctx.Background(), owner, repo, sha, input)
	if err != nil {
		return err
	}
	// We could see a 400 error and in that case we should not try to react to the comment
	if !isError && response.StatusCode == 201 {
		if _, _, err = gn.client.Reactions.CreateCommentReaction(ctx.Background(), owner, repo, *comment.ID, "rocket"); err != nil {
			return err
		}
	}
	return nil
}

func buildFirstComment(body, message string, isError bool) string {
	var output bytes.Buffer
	if isError {
		output.WriteString(fmt.Sprintf("#### Failed\n> :x: &nbsp; %s\n", message))
	} else {
		output.WriteString(fmt.Sprintf("#### Successful\n> :white_check_mark: &nbsp; %s\n", message))
	}
	output.WriteString("<details>\n")
	output.WriteString("<summary>Logs</summary>\n")
	output.WriteString("\n")
	output.WriteString("#### Detail\n")
	output.WriteString("\n")
	output.WriteString("```")
	output.WriteString("\n")
	output.WriteString(body)
	output.WriteString("\n")
	output.WriteString("```")
	output.WriteString("\n")
	output.WriteString("<div align=\"right\">\n")
	output.WriteString("<img src=\"https://avatars1.githubusercontent.com/u/20845599?s=200&v=4\" width=\"20px\" />")
	output.WriteString("<a href=\"https://docs.armory.io/docs/spinnaker-user-guides/using-dinghy\" target=\"_blank\" > Pipeline as Code</a> ")
	output.WriteString("</div>\n")
	output.WriteString("</details>\n")
	return output.String()
}

func buildSubsequentComments(body string) string {
	var output bytes.Buffer
	output.WriteString("<details>\n")
	output.WriteString("<summary>Additional Logs</summary>\n")
	output.WriteString("\n")
	output.WriteString("#### Detail\n")
	output.WriteString("\n")
	output.WriteString("```")
	output.WriteString("\n")
	output.WriteString(body)
	output.WriteString("\n")
	output.WriteString("```")
	output.WriteString("\n")
	output.WriteString("<div align=\"right\">\n")
	output.WriteString("<img src=\"https://avatars1.githubusercontent.com/u/20845599?s=200&v=4\" width=\"20px\" />")
	output.WriteString("<a href=\"https://docs.armory.io/docs/spinnaker-user-guides/using-dinghy\" target=\"_blank\" > Pipeline as Code</a> ")
	output.WriteString("</div>\n")
	output.WriteString("</details>\n")
	return output.String()
}
