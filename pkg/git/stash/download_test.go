package stash

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/armory-io/dinghy/pkg/settings"
)

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
			repo:        "myrepo",
			path:        "my/path.yml",
			url:         "/projects/armory/repos/myrepo/browse/my/path.yml?raw",
		},
	}

	downloader := &FileService{}
	for _, c := range cases {
		c.setEndpoint()
		org, repo, path := downloader.DecodeURL(c.url)

		assert.Equal(t, c.owner, org)
		assert.Equal(t, c.repo, repo)
		assert.Equal(t, c.path, path)
	}
}
