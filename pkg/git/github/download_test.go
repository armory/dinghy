package github

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/armory-io/dinghy/pkg/settings"
)

func TestEncodeUrl(t *testing.T) {
	cases := []struct {
		setEndpoint func()
		owner       string
		repo        string
		path        string
		expected    string
	}{
		{
			setEndpoint: func() { settings.S.GithubEndpoint = "https://api.github.com" },
			owner:       "armory",
			repo:        "armory",
			path:        "my/path.yml",
			expected:    "https://api.github.com/repos/armory/armory/contents/my/path.yml",
		},
		{
			setEndpoint: func() { settings.S.GithubEndpoint = "https://mygithub.com" },
			owner:       "armory",
			repo:        "armory",
			path:        "my/path.yml",
			expected:    "https://mygithub.com/repos/armory/armory/contents/my/path.yml",
		},
	}

	downloader := &FileService{}
	for _, c := range cases {
		c.setEndpoint()
		if out := downloader.EncodeURL(c.owner, c.repo, c.path); out != c.expected {
			t.Errorf("%s did not match expected %s", out, c.expected)
		}
	}
}

func TestDecodeUrl(t *testing.T) {
	cases := []struct {
		setEndpoint func()
		owner       string
		repo        string
		path        string
		url         string
	}{
		{
			setEndpoint: func() { settings.S.GithubEndpoint = "https://api.github.com" },
			owner:       "armory",
			repo:        "armory",
			path:        "my/path.yml",
			url:         "https://api.github.com/repos/armory/armory/contents/my/path.yml",
		},
		{
			setEndpoint: func() { settings.S.GithubEndpoint = "https://mygithub.com" },
			owner:       "armory",
			repo:        "armory",
			path:        "my/path.yml",
			url:         "https://mygithub.com/repos/armory/armory/contents/my/path.yml",
		},
	}

	downloader := &FileService{}
	for _, c := range cases {
		c.setEndpoint()
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
	expected := "File data"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, expected)
	}))
	defer ts.Close()
	// TODO: Do not use global variable. This will lead to side-effects.
	settings.S.GithubEndpoint = ts.URL
	fs := FileService{}
	data, err := fs.Download("org", "repo", "path")
	assert.Nil(t, err)
	assert.Equal(t, expected, data)
	// Verify cache usage:
	assert.Equal(t, 1, fs.cache.Len())
	v := fs.cache.Get(fs.EncodeURL("org", "repo", "path"))
	assert.Equal(t, expected, v)
}
