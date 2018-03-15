package dinghyfile

import (
	"strings"
	"testing"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/git/github"
	"github.com/stretchr/testify/assert"
)

const simpleTempl = `
{
    "stages": [
        {{ module "wait.stage.module" "waitTime" 10 }}
    ],
}
`

const expected = `
{
    "stages": [
        {
            "name": "Wait",
            "refId": "1",
            "requisiteStageRefIds": [],
            "type": "wait",
            "waitTime": 10
        }
    ],
}
`

// FileService is for working with repositories
type FileService struct{}

// Download a file from github.
func (f *FileService) Download(org, repo, file string) (string, error) {
	ret := `{
        "name": "Wait",
        "refId": "1",
        "requisiteStageRefIds": [],
        "type": "wait",
        "waitTime": 12044
    }`
	return ret, nil
}

// GitURL returns the git url for a given org, repo, path
func (f *FileService) GitURL(org, repo, path string) string {
	return (&github.FileService{}).GitURL(org, repo, path)
}

// ParseGitURL takes a url and returns the org, repo, path
func (f *FileService) ParseGitURL(url string) (org, repo, path string) {
	return (&github.FileService{}).ParseGitURL(url)
}

func TestSimpleWaitStage(t *testing.T) {
	buf := Render(cache.NewCache(), "simpleTempl", simpleTempl, "org", "repo", &FileService{})

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}
