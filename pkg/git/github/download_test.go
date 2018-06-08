package github

import (
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
