package notifiers

import (
	ctx "context"
	"net/url"
	"reflect"
	"regexp"
	"testing"

	"github.com/armory-io/dinghy/pkg/settings"
	ossSettings "github.com/armory/dinghy/pkg/settings/global"
	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func Test_newGithubNotification(t *testing.T) {
	type args struct {
		owner   string
		repo    string
		content map[string]interface{}
		isError bool
		message string
	}
	tests := []struct {
		name string
		args args
		want []GithubNotification
	}{
		{
			name: "GithubNotifications should be instantiated",
			args: args{
				owner:   "",
				repo:    "",
				content: map[string]interface{}{},
				isError: false,
				message: "",
			},
			want: generateNotifications("", "", map[string]interface{}{}, false, ""),
		},
		{
			name: "GithubNotification should be instantiated with nil content",
			args: args{
				owner:   "",
				repo:    "",
				content: nil,
				isError: false,
			},
			want: generateNotifications("", "", nil, false, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateNotifications(tt.args.owner, tt.args.repo, tt.args.content, tt.args.isError, tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newGithubNotification() = %v, want %v", got, tt.want)
			}
		})
	}
}

type Helper struct {
	*GithubNotifier
	Error error
}

func helper(token, baseURL string, e error) Helper {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx.Background(), ts)

	c := github.NewClient(tc)
	u, _ := url.Parse(baseURL)
	c.BaseURL = u

	return Helper{
		GithubNotifier: &GithubNotifier{client: c},
		Error:          e,
	}
}

func TestNewGithubNotifier(t *testing.T) {
	type args struct {
		s *settings.ExtSettings
	}
	tests := []struct {
		name string
		args args
		want Helper
	}{
		{
			name: "GithubNotifier should be instantiated",
			args: args{
				s: &settings.ExtSettings{
					Settings: &ossSettings.Settings{
						GitHubToken: "mytoken",
					},
				},
			},
			want: helper("mytoken", PublicGithubEndpoint+"/", nil),
		},
		{
			name: "GithubNotifier should be instantiated with empty GithubToken",
			args: args{
				s: &settings.ExtSettings{
					Settings: &ossSettings.Settings{
						GitHubToken: "",
					},
				},
			},
			want: helper("", PublicGithubEndpoint+"/", nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGithubNotifier(tt.args.s)
			if tt.want.Error != nil && err != tt.want.Error {
				t.Errorf("Errors did not match. expected: %v actual: %v", tt.want.Error, err)
			} else if !reflect.DeepEqual(got, tt.want.GithubNotifier) {
				t.Errorf("NewGithubNotifier() = %v, want %v", got, tt.want.GithubNotifier)
			}
		})
	}
}

func Test_buildComment(t *testing.T) {
	type args struct {
		body    string
		message string
		isError bool
	}
	tests := []struct {
		name          string
		args          args
		shouldContain string
	}{
		{
			name: "Posted message should contain message",
			args: args{
				message: "mymessage",
				isError: false,
			},
			shouldContain: "mymessage",
		},
		{
			name: "Posted message should contain body",
			args: args{
				body:    "mybody",
				isError: false,
			},
			shouldContain: "mybody",
		},
		{
			name: "Posted message should contain fail title",
			args: args{
				isError: true,
			},
			shouldContain: "Failed",
		},
		{
			name: "Posted message should contain success title",
			args: args{
				isError: false,
			},
			shouldContain: "Successful",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFirstComment(tt.args.body, tt.args.message, tt.args.isError)
			var re = regexp.MustCompile(tt.shouldContain)
			assert.Equal(t, true, re.MatchString(got))
		})
	}
}
