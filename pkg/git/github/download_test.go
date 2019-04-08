package github

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		expected string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			expected: "https://api.github.com/repos/armory/armory/contents/my/path.yml",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			expected: "https://mygithub.com/repos/armory/armory/contents/my/path.yml",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		if out := downloader.EncodeURL(c.owner, c.repo, c.path); out != c.expected {
			t.Errorf("%s did not match expected %s", out, c.expected)
		}
	}
}

func TestDecodeUrl(t *testing.T) {
	cases := []struct {
		endpoint string
		owner    string
		repo     string
		path     string
		url      string
	}{
		{
			endpoint: "https://api.github.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			url:      "https://api.github.com/repos/armory/armory/contents/my/path.yml",
		},
		{
			endpoint: "https://mygithub.com",
			owner:    "armory",
			repo:     "armory",
			path:     "my/path.yml",
			url:      "https://mygithub.com/repos/armory/armory/contents/my/path.yml",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			GitHub: &GitHubTest{
				endpoint: c.endpoint,
			},
		}
		org, repo, path := downloader.DecodeURL(c.url)

		if org != c.owner {
			t.Errorf("%s did not match expected owner %s", org, c.owner)
		}

		if repo != c.repo {
			t.Errorf("%s did not match expected repo %s", repo, c.repo)
		}

		if path != c.path {
			t.Errorf("%s did not match expected path %s", path, c.path)
		}
	}
}

func TestDownload(t *testing.T) {
	testCases := map[string]struct {
		org         string
		repo        string
		path        string
		fs          *FileService
		expected    string
		expectedErr error
	}{
		"success": {
			org:  "org",
			repo: "repo",
			path: "path",
			fs: &FileService{
				GitHub: &GitHubTest{contents: "file contents"},
			},
			expected:    "file contents",
			expectedErr: nil,
		},
		"error": {
			org:  "org",
			repo: "repo",
			path: "path",
			fs: &FileService{
				GitHub: &GitHubTest{
					contents: "",
					err:      errors.New("fail"),
				},
			},
			expected:    "",
			expectedErr: errors.New("fail"),
		},
	}

	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			actual, err := tc.fs.Download(tc.org, tc.repo, tc.path)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.expectedErr, err)

			// test caching
			v := tc.fs.cache.Get(tc.fs.EncodeURL("org", "repo", "path"))
			assert.Equal(t, tc.expected, v)
		})
	}
}
