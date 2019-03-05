package stash

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
			repo:     "myrepo",
			path:     "my/path.yml",
			url:      "/projects/armory/repos/myrepo/browse/my/path.yml?raw",
		},
	}

	for _, c := range cases {
		downloader := &FileService{
			StashEndpoint: c.endpoint,
		}
		org, repo, path := downloader.DecodeURL(c.url)

		assert.Equal(t, c.owner, org)
		assert.Equal(t, c.repo, repo)
		assert.Equal(t, c.path, path)
	}
}
