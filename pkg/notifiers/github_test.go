package notifiers

import (
	"github.com/armory-io/dinghy/pkg/settings"
	ossSettings "github.com/armory/dinghy/pkg/settings"
	"github.com/stretchr/testify/assert"
	"reflect"
	"regexp"
	"testing"
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
		want *GithubNotification
	}{
		{
			name: "GithubNotification should be instantiated",
			args: args{
				owner:   "",
				repo:    "",
				content: map[string]interface{}{},
				isError: false,
				message: "",
			},
			want: newGithubNotification("", "", map[string]interface{}{}, false, ""),
		},
		{
			name: "GithubNotification should be instantiated with nil content",
			args: args{
				owner:   "",
				repo:    "",
				content: nil,
				isError: false,
			},
			want: newGithubNotification("", "", nil, false, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newGithubNotification(tt.args.owner, tt.args.repo, tt.args.content, tt.args.isError, tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newGithubNotification() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGithubNotifier(t *testing.T) {
	type args struct {
		s *settings.ExtSettings
	}
	tests := []struct {
		name string
		args args
		want *GithubNotifier
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
			want: NewGithubNotifier(&settings.ExtSettings{
				Settings: &ossSettings.Settings{
					GitHubToken: "mytoken",
				},
			}),
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
			want: NewGithubNotifier(&settings.ExtSettings{
				Settings: &ossSettings.Settings{
					GitHubToken: "",
				},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGithubNotifier(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGithubNotifier() = %v, want %v", got, tt.want)
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
			got := buildComment(tt.args.body, tt.args.message, tt.args.isError)
			var re = regexp.MustCompile(tt.shouldContain)
			assert.Equal(t, true, re.MatchString(got))
		})
	}
}
