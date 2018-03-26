package dinghyfile

import (
	"strings"
	"testing"

	"encoding/json"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/git/github"
	"github.com/stretchr/testify/assert"
)

const simpleTempl = `
{
    "stages": [
        {{ module "wait.stage.module" "waitTime" 10 "refId" { "c": "d" } "requisiteStageRefIds" ["1", "2", "3"] }}
    ],
}
`

const expected = `
{
    "stages": [
        {
            "name": "Wait",
            "refId": { "c": "d" },
            "requisiteStageRefIds": ["1", "2", "3"],
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
        "refId": {},
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

type MultilevelFileService struct{}

// Download a file from github.
func (f *MultilevelFileService) Download(org, repo, file string) (string, error) {
	switch file {
	case "dinghyfile":
		return `{{ module "wait.stage.module" "foo" "baz" }}`, nil

	case "wait.stage.module":
		return `{
			"foo": "bar",
			"nested": {{ module "wait.dep.module" "waitTime" 100 }}
		}`, nil

	case "wait.dep.module":
		return `{
			"waitTime": 4243
		}`, nil
	}

	return "", nil
}

// GitURL returns the git url for a given org, repo, path
func (f *MultilevelFileService) GitURL(org, repo, path string) string {
	return (&github.FileService{}).GitURL(org, repo, path)
}

// ParseGitURL takes a url and returns the org, repo, path
func (f *MultilevelFileService) ParseGitURL(url string) (org, repo, path string) {
	return (&github.FileService{}).ParseGitURL(url)
}

func TestModuleVariableSubstitution(t *testing.T) {
	cache.C = cache.NewCache()
	f := MultilevelFileService{}

	file, err := f.Download("org", "repo", "dinghyfile")
	assert.Equal(t, nil, err)

	ret := Render(cache.C, "dinghyfile", file, "org", "repo", &f)

	type testStruct struct {
		Foo    string `json:"foo"`
		Nested struct {
			WaitTime int `json:"waitTime"`
		} `json:"nested"`
	}

	var ts = testStruct{}
	err = json.Unmarshal(ret.Bytes(), &ts)
	assert.Equal(t, nil, err)
	assert.Equal(t, "baz", ts.Foo)
	assert.Equal(t, 100, ts.Nested.WaitTime)
}
