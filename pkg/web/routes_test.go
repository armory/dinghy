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

package web

import (
	"bytes"
	"errors"
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/armory/dinghy/pkg/git/github"
	dinghylog "github.com/armory/dinghy/pkg/log"
	"github.com/armory/dinghy/pkg/logevents"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/armory/dinghy/pkg/util"
	"github.com/sirupsen/logrus"

	// "errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/armory/dinghy/pkg/mock"

	// "github.com/armory/dinghy/pkg/settings"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRouterSanity(t *testing.T) {
	wa := &WebAPI{}
	wa.MetricsHandler = new(NoOpMetricsHandler)
	r := wa.Router(new(global.Settings))
	assert.Equal(t, "*mux.Router", reflect.TypeOf(r).String())
}

func testHealthCheckLogging(t *testing.T, endpoint string) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Debug(gomock.Any()).Times(1)
	logger.EXPECT().Info(gomock.Any()).Times(0)

	wa := NewWebAPI(nil, nil, nil, logger, nil, nil, nil, nil)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(wa.healthcheck)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"status":"ok"}`, rr.Body.String())
}

func TestHealthCheckLogging(t *testing.T) {
	// This is served by multiple endpoints, so we'll iterate over all of them.
	endpoints := []string{"/", "/health", "/healthcheck"}
	for _, e := range endpoints {
		testHealthCheckLogging(t, e)
	}
}

// Github webhook tests
func TestGithubWebhookHandlerBadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode github webhook: %s"), gomock.Any()).Times(1)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})

	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)

	payload := bytes.NewBufferString(`{broken`)
	req := httptest.NewRequest("POST", "/v1/webhooks/github", payload)
	rr := httptest.NewRecorder()
	wa.githubWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestGithubWebhookHandlerNoRef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Info(gomock.Eq(stringToInterfaceSlice("Possibly a non-Push notification received (blank ref)"))).Times(1)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})
	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)

	payload := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest("POST", "/v1/webhooks/github", payload)
	rr := httptest.NewRecorder()
	wa.githubWebhookHandler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGithubWebhookHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(5)
	logger.EXPECT().Info(gomock.Any()).Times(2)
	logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).Times(1)
	logger.EXPECT().WithFields(gomock.Any())
	logger.EXPECT().Error(gomock.Any()).Times(1)

	s := source.NewMockSourceConfiguration(ctrl)
	s.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{
			TemplateRepo:                      "my-repo",
			TemplateOrg:                       "my-org",
			GitHubToken:                       "GitHubToken",
			DinghyFilename:                    "dinghyfile",
			AutoLockPipelines:                 "true",
			WebhookValidationEnabledProviders: []string{"github"},
			WebhookValidations:                []global.WebhookValidation{},
			SpinnakerSupplied: global.SpinnakerSupplied{
				Deck: global.SpinnakerService{
					Enabled: "",
					BaseURL: "",
				},
			},
			RepoConfig:                  []global.RepoConfig{},
			RepositoryRawdataProcessing: true,
		}, dinghyfile.NewMockPlankClient(ctrl), nil
	})

	wa := NewWebAPI(s, nil, nil, logger, nil, nil, nil, nil)
	wa.Parser = dinghyfile.NewDinghyfileParser(&dinghyfile.PipelineBuilder{})
	payload := bytes.NewBufferString(`{"ref":"refs/heads/some_branch","repository":{"name":"my-repo","organization":"my-org"}}`)
	req := httptest.NewRequest("POST", "/v1/webhooks/github", payload)
	rr := httptest.NewRecorder()
	wa.githubWebhookHandler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"status":"accepted"}`, rr.Body.String())

}

// This function is needed for the not formatted messages after dinghylog implementation
func stringToInterfaceSlice(args ...interface{}) []interface{} {
	return args
}

// Legacy Bitbucket ("Stash") webhook tests
func TestStashWebhookHandlerBadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Reading stash payload body"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode stash webhook: %s"), gomock.Any()).Times(1)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})
	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)

	payload := bytes.NewBufferString(`{broken`)
	req := httptest.NewRequest("POST", "/v1/webhooks/stash", payload)
	rr := httptest.NewRecorder()
	wa.stashWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestStashWebhookBadPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Infof(gomock.Eq("Reading stash payload body"), gomock.Any()).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode stash webhook: %s"), gomock.Any()).Times(1)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})

	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)
	payload := bytes.NewBufferString(`{"event_type": "stash", "refChanges": "not an array"}`)

	req := httptest.NewRequest("POST", "/v1/webhooks/stash", payload)
	rr := httptest.NewRecorder()
	wa.stashWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

// Bitbucket Server webhook tests
func TestBitbucketWebhookHandlerBadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Unable to determine bitbucket event type: %s"), gomock.Any()).Times(1)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})

	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)
	payload := bytes.NewBufferString(`{broken`)
	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestBitbucketWebhookBadPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Info(gomock.Eq(stringToInterfaceSlice("Processing bitbucket-server webhook"))).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode bitbucket-server webhook: %s"), gomock.Any()).Times(1)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})

	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)

	payload := bytes.NewBufferString(`{"event_type": "repo:refs_changed", "changes": "not an array"}`)

	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

// Bitbucket Cloud webhook tests
func TestBitbucketCloudWebhookHandlerBadJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Unable to determine bitbucket event type: %s"), gomock.Any()).Times(1)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})

	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)

	payload := bytes.NewBufferString(`{broken`)
	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket-cloud", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestBitbucketCloudWebhookBadPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().Info(gomock.Eq(stringToInterfaceSlice("Processing bitbucket-cloud webhook"))).Times(1)
	logger.EXPECT().Infof(gomock.Eq("Received payload: %s"), gomock.Any()).Times(2)
	logger.EXPECT().Errorf(gomock.Eq("failed to decode bitbucket-cloud webhook: %s"), gomock.Any()).Times(1)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})

	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)

	payload := bytes.NewBufferString(`{"event_type": "repo:push", "push": {"changes": "not an array"}}`)

	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket-cloud", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 422, rr.Code)
}

func TestUnknownEventType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	logger.EXPECT().WithFields(gomock.Any())

	sc := source.NewMockSourceConfiguration(ctrl)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return &global.Settings{}, dinghyfile.NewMockPlankClient(ctrl), nil
	})

	wa := NewWebAPI(sc, nil, nil, logger, nil, nil, nil, nil)

	payload := bytes.NewBufferString(`{"event_type": "", "changes": "not an array"}`)

	req := httptest.NewRequest("POST", "/v1/webhooks/bitbucket-cloud", payload)
	rr := httptest.NewRecorder()
	wa.bitbucketWebhookHandler(rr, req)
	assert.Equal(t, 500, rr.Code)
}

func TestLogevents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)

	lec := logevents.NewMockLogEventsClient(ctrl)
	lec.EXPECT().GetLogEvents().AnyTimes().DoAndReturn(func() ([]logevents.LogEvent, error) {
		event := logevents.LogEvent{
			Org:  "org",
			Repo: "repo",
			Files: []string{
				"file1",
			},
			Message: "",
			Date:    0,
			Commits: []string{
				"12345",
			},
			Status:             "mystatus",
			RawData:            "raw",
			RenderedDinghyfile: "dinghyfile",
			PullRequest:        "https://github/pr",
		}
		return []logevents.LogEvent{event}, nil
	})

	wa := NewWebAPI(nil, nil, nil, logger, nil, nil, lec, nil)
	req, err := http.NewRequest("GET", "/v1/logevents", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(wa.logevents)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `[{"org":"org","repo":"repo","files":["file1"],"message":"","date":0,"commits":["12345"],"status":"mystatus","rawdata":"raw","rendereddinghyfile":"dinghyfile","pullrequest":"https://github/pr"}]`, rr.Body.String())
}

func Test_contains(t *testing.T) {
	type args struct {
		whvalidations []string
		provider      string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should return false, since provided whvalidations are nil",
			args: args{
				whvalidations: nil,
				provider:      "github",
			},
			want: false,
		},
		{
			name: "Should return false, since provider is not within provided whvalidations",
			args: args{
				whvalidations: []string{
					"gitlab",
				},
				provider: "github",
			},
			want: false,
		},
		{
			name: "Should return true, since provider exist within provided whvalidations",
			args: args{
				whvalidations: []string{
					"github",
				},
				provider: "github",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.whvalidations, tt.args.provider); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getWebhookSecret(t *testing.T) {

	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should be an empty string, since the requested header is not present within http.Request",
			args: args{
				r: func() *http.Request {
					req, err := http.NewRequest("GET", "", nil)
					if err != nil {
						t.Fatal(err)
					}
					return req
				}(),
			},
			want: "",
		},
		{
			name: "Should be not an empty string, since the requested header is present within http.Request",
			args: args{
				r: func() *http.Request {
					req, err := http.NewRequest("GET", "", nil)
					req.Header.Add("webhook-secret", "mysecret")
					if err != nil {
						t.Fatal(err)
					}
					return req
				}(),
			},
			want: "mysecret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHeader(tt.args.r, "webhook-secret"); got != tt.want {
				t.Errorf("getWebhookSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRawPayload(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "RawPayload is present, since it is present within JSON payload",
			args: args{
				body: []byte(`{"raw_payload":"raw"}`),
			},
			want: "raw",
		},
		{
			name: "RawPayload is not present, since it is not present within JSON payload",
			args: args{
				body: []byte(`{"key":"value"}`),
			},
			want: "",
		},
		{
			name: "RawPayload is not present, since JSON payload is invalid",
			args: args{
				body: []byte(``),
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRawPayload(tt.args.body); got != tt.want {
				t.Errorf("getRawPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldRunValidationWhenBranchIsCorrect(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	dl := dinghylog.NewDinghyLogs(l)

	p := github.Push{
		Logger: dl,
		Repository: github.Repository{
			Name: "test_repo",
		},
		Ref: "test_branch",
	}
	s := &global.Settings{
		DinghyFilename: "dinghyfile",
		TemplateOrg:    "test_org",
		TemplateRepo:   "test_repo",
		RepoConfig: []global.RepoConfig{
			{
				Provider: "github",
				Repo:     "test_repo",
				Branch:   "test_branch",
			},
		},
		GitHubToken: "test_github_token",
	}

	assert.False(t, shouldRunValidation(&p, s, dl))
}

func TestShouldRunValidationWhenBranchIsWrongAndIsNotMaster(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Infof(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	dl := dinghylog.NewDinghyLogs(l)

	p := github.Push{
		Logger: dl,
		Repository: github.Repository{
			Name: "test_repo",
		},
		Ref: "wrong_branch",
	}
	s := &global.Settings{
		DinghyFilename: "dinghyfile",
		TemplateOrg:    "test_org",
		TemplateRepo:   "test_repo",
		RepoConfig: []global.RepoConfig{
			{
				Provider: "github",
				Repo:     "test_repo",
				Branch:   "test_branch",
			},
		},
		GitHubToken: "test_github_token",
	}

	assert.True(t, shouldRunValidation(&p, s, dl))
}

func TestShouldRunValidationWhenBranchIsWrongAndIsMaster(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(1)

	dl := dinghylog.NewDinghyLogs(l)

	p := github.Push{
		Logger: dl,
		Repository: github.Repository{
			Name: "test_repo",
		},
		Ref: "master",
	}
	s := &global.Settings{
		MultipleBranchesEnabled: "true",
		DinghyFilename:          "dinghyfile",
		TemplateOrg:             "test_org",
		TemplateRepo:            "test_repo",
		RepoConfig: []global.RepoConfig{
			{
				Provider: "github",
				Repo:     "test_repo",
				Branch:   "test_branch",
			},
		},
		GitHubToken: "test_github_token",
	}

	assert.False(t, shouldRunValidation(&p, s, dl))
}

func TestGetIgnoreFilePatternsWhenDinghyIgnoreFetched(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(1)

	dl := dinghylog.NewDinghyLogs(l)
	p := github.Push{
		Logger: dl,
		Repository: github.Repository{
			Organization: "test_org",
			Name:         "test_repo",
		},
		Ref: "test_branch",
	}

	f := dinghyfile.NewMockDownloader(c)
	f.EXPECT().Download(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("ab\ncd", nil)

	result := getIgnoreFilePatterns(&p, f, dl)

	assert.Equal(t, result[0], "ab")
	assert.Equal(t, result[1], "cd")
}

func TestGetIgnoreFilePatternsWhenDinghyIgnoreFetchingFailed(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Info(gomock.Any()).Times(1)

	dl := dinghylog.NewDinghyLogs(l)
	p := github.Push{
		Logger: dl,
		Repository: github.Repository{
			Organization: "test_org",
			Name:         "test_repo",
		},
		Ref: "test_branch",
	}

	f := dinghyfile.NewMockDownloader(c)
	f.EXPECT().Download(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("something went wrong"))

	result := getIgnoreFilePatterns(&p, f, dl)

	assert.Equal(t, len(result), 0)
}

func TestBuildPipelinesWhenDinghyIgnoreRegexp2Enabled(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Info(gomock.Any()).Times(1)
	l.EXPECT().Error(gomock.Any()).Times(2)
	l.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(6)
	l.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	dl := dinghylog.NewDinghyLogs(l)

	p := github.Push{
		Logger: dl,
		Repository: github.Repository{
			Organization: "test_org",
			Name:         "test_repo",
		},
		Commits: []github.Commit{
			{Added: []string{"file.js", "file.css", "file.html"}},
		},
		Ref: "test_branch",
	}
	s := &global.Settings{
		DinghyFilename: "dinghyfile",
		TemplateOrg:    "test_org",
		TemplateRepo:   "test_repo",
		RepoConfig: []global.RepoConfig{
			{
				Provider: "github",
				Repo:     "test_repo",
				Branch:   "test_branch",
			},
		},
		GitHubToken:                "test_github_token",
		DinghyIgnoreRegexp2Enabled: "true",
	}

	sc := source.NewMockSourceConfiguration(c)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return s, dinghyfile.NewMockPlankClient(c), nil
	})

	wa := NewWebAPI(sc, nil, nil, l, nil, nil, nil, nil)
	wa.Parser = dinghyfile.NewDinghyfileParser(&dinghyfile.PipelineBuilder{})

	r := httptest.NewRecorder()

	d := dinghyfile.NewMockDownloader(c)
	d.EXPECT().Download("test_org", "test_repo", ".dinghyignore", "test_branch").Return("file.(js|css|html)", nil)

	wa.buildPipelines(&p, []byte("{}"), d, r, dl, "", nil, s)

	assert.Equal(t, http.StatusOK, r.Code)
	assert.Equal(t, `{"status":"accepted"}`, r.Body.String())
}

func TestBuildPipelinesWhenDinghyIgnoreRegexp2Disabled(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l := mock.NewMockFieldLogger(c)
	l.EXPECT().Info(gomock.Any()).Times(1)
	l.EXPECT().Error(gomock.Any()).Times(2)
	l.EXPECT().Infof(gomock.Any(), gomock.Any()).Times(6)
	l.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	dl := dinghylog.NewDinghyLogs(l)

	p := github.Push{
		Logger: dl,
		Repository: github.Repository{
			Organization: "test_org",
			Name:         "test_repo",
		},
		Commits: []github.Commit{
			{Added: []string{"file.js", "file.css", "file.html"}},
		},
		Ref: "test_branch",
	}
	s := &global.Settings{
		DinghyFilename: "dinghyfile",
		TemplateOrg:    "test_org",
		TemplateRepo:   "test_repo",
		RepoConfig: []global.RepoConfig{
			{
				Provider: "github",
				Repo:     "test_repo",
				Branch:   "test_branch",
			},
		},
		GitHubToken:                "test_github_token",
		DinghyIgnoreRegexp2Enabled: "false",
	}

	sc := source.NewMockSourceConfiguration(c)
	sc.EXPECT().GetSettings(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(r *http.Request, logger2 *logrus.Logger) (*global.Settings, util.PlankClient, error) {
		return s, dinghyfile.NewMockPlankClient(c), nil
	})

	wa := NewWebAPI(sc, nil, nil, l, nil, nil, nil, nil)
	wa.Parser = dinghyfile.NewDinghyfileParser(&dinghyfile.PipelineBuilder{})

	r := httptest.NewRecorder()

	d := dinghyfile.NewMockDownloader(c)
	d.EXPECT().Download("test_org", "test_repo", ".dinghyignore", "test_branch").Return("file.(js|css|html)", nil)

	wa.buildPipelines(&p, []byte("{}"), d, r, dl, "", nil, s)

	assert.Equal(t, http.StatusOK, r.Code)
	assert.Equal(t, `{"status":"accepted"}`, r.Body.String())
}
