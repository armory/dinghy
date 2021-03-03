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
	"errors"
	"github.com/armory/dinghy/pkg/cache/local"
	_ "github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/armory/dinghy/pkg/log"
	"github.com/armory/dinghy/pkg/mock"
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestEncodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		branch   string
		expected string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			expected: "https://api.github.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			expected: "https://mygithub.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		actual := downloader.EncodeURL(c.owner, c.repo, c.path, c.branch)
		assert.Equal(t, c.expected, actual)
	}
}

func TestEncodeUrlWithLeadingSlashs(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		branch   string
		expected string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "/my/path.yml",
			branch:   "mybranch",
			expected: "https://api.github.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			expected: "https://mygithub.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		actual := downloader.EncodeURL(c.owner, c.repo, c.path, c.branch)
		assert.Equal(t, c.expected, actual)
	}
}

func TestDecodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		branch   string
		url      string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			url:      "https://api.github.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			branch:   "mybranch",
			url:      "https://mygithub.com/repos/armory/armory/contents/my/path.yml?ref=mybranch",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		org, repo, path, branch := downloader.DecodeURL(c.url)
		assert.Equal(t, c.owner, org)
		assert.Equal(t, c.repo, repo)
		assert.Equal(t, c.path, path)
		assert.Equal(t, c.branch, branch)
	}
}

func TestDownload(t *testing.T) {
	testCases := map[string]struct {
		org         string
		repo        string
		path        string
		branch      string
		fs          *FileService
		contains    string
		expectedErr error
	}{
		"success": {
			org:    "armory",
			repo:   "dinghy",
			path:   "example/dinghyfile",
			branch: "master",
			fs: &FileService{
				GitHub: &GitHubTest{endpoint: "https://api.github.com", contents: "file contents"},
				Logger: log.DinghyLogs{Logs: map[string]log.DinghyLogStruct{
					log.SystemLogKey: {
						Logger:         logrus.New(),
						LogEventBuffer: &bytes.Buffer{},
					},
				}},
			},
			contains:    "dinghy",
			expectedErr: nil,
		},
		"error": {
			org:    "org",
			repo:   "repo",
			path:   "path",
			branch: "branch",
			fs: &FileService{
				GitHub: &GitHubTest{
					endpoint: "https://api.github.com",
					contents: "",
					err:      errors.New("fail"),
				},
				Logger: log.DinghyLogs{Logs: map[string]log.DinghyLogStruct{
					log.SystemLogKey: {
						Logger:         logrus.New(),
						LogEventBuffer: &bytes.Buffer{},
					},
				}},
			},
			contains:    "",
			expectedErr: errors.New("File path not found for org org in repository repo"),
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual, err := tc.fs.Download(tc.org, tc.repo, tc.path, tc.branch)
			assert.Contains(t, actual, tc.contains)
			if tc.expectedErr == nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			}

			// test caching
			v := tc.fs.cache.Get(tc.fs.EncodeURL("org", "repo", "path", "branch"))
			assert.Contains(t, tc.contains, v)
		})
	}
}

func TestMasterDownloadFailsAndTriesMain(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mock.NewMockFieldLogger(ctrl)
	fs := &FileService{
		GitHub: &GitHubTest{contents: "", err: errors.New("fail")},
		Logger: log.DinghyLogs{Logs: map[string]log.DinghyLogStruct{
			log.SystemLogKey: {
				Logger:         logger,
				LogEventBuffer: &bytes.Buffer{},
			},
		}},
	}

	logger.EXPECT().Error(gomock.Any()).Times(2)
	logger.EXPECT().Info(gomock.Eq(stringToSlice("DownloadContents failed with master branch, trying with main branch"))).Times(1)
	logger.EXPECT().Errorf(gomock.Eq("Download failed also for branch %v"), gomock.Any()).Times(1)

	fs.Download("org", " repo", "path", "master")
}

func stringToSlice(args ...interface{}) []interface{} {
	return args
}

func TestFileService_DownloadContents(t *testing.T) {
	type fields struct {
		cache  local.Cache
		GitHub GitHubClient
		Logger log.DinghyLog
	}
	type args struct {
		org    string
		repo   string
		path   string
		branch string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		contains string
		wantErr  bool
	}{
		{
			name: "Content should be downloaded correctly",
			fields: fields{
				GitHub: &GitHubTest{endpoint: "https://api.github.com"},
			},
			args: args{
				org:    "armory",
				repo:   "dinghy",
				path:   "example/dinghyfile",
				branch: "master",
			},
			wantErr: false,
		},
		{
			name: "Large files should be downloaded correctly",
			fields: fields{
				GitHub: &GitHubTest{endpoint: "https://api.github.com"},
			},
			args: args{
				org:    "armory",
				repo:   "se-pipeline-files",
				path:   "kubernetesdemo/dinghyfile",
				branch: "master",
			},
			wantErr: false,
		},
		{
			name: "Download should fail, file not exist.",
			fields: fields{
				GitHub: &GitHubTest{endpoint: "https://api.github.com"},
			},
			args: args{
				org:    "example",
				repo:   "test",
				path:   "dinghyfile",
				branch: "master",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileService{
				cache:  tt.fields.cache,
				GitHub: tt.fields.GitHub,
				Logger: tt.fields.Logger,
			}
			_, err := f.DownloadContents(tt.args.org, tt.args.repo, tt.args.path, tt.args.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
