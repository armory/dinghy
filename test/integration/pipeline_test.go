package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/git/status"
)

const dinghf = `{
	"application": "dinghyintegration",
	"pipelines": [
	  {
		"application": "dinghyintegration",
		"keepWaitingPipelines": false,
		"limitConcurrent": false,
		"name": "This is new",
		"stages": [
			{{ module "wait.stage.module" "waitTime" 10 }}
		],
		"triggers": []
	  }
	]
}`

const mod = `{
    "name": "Wait",
    "refId": "1",
    "requisiteStageRefIds": [],
    "type": "wait",
    "waitTime": 1000
}`

// FileService is for working with repositories
type FileService struct{}

// Download a file from github.
func (f *FileService) Download(org, repo, file string) (string, error) {
	if file == "dinghyfile" {
		return dinghf, nil
	}
	return mod, nil
}

// Push is for a github push notification
type Push struct{}

// ContainsFile checks to see if a given file is in the push.
func (p *Push) ContainsFile(file string) bool {
	return true
}

// IsMaster checks if ref is master
func (p *Push) IsMaster() bool {
	return true
}

// Files returns the list of files modified
func (p *Push) Files() []string {
	return []string{"foo"}
}

// Repo returns the repo
func (p *Push) Repo() string {
	return "dinghyintegration"
}

// Org returns the org
func (p *Push) Org() string {
	return "armory-io"
}

// SetCommitStatus sets the commit status
func (p *Push) SetCommitStatus(s status.Status) {
}

// TestSpinnakerPipelineUpdate tests pipeline update in spinnaker
// even though it mocks out the github part, it talks to spinnaker
// hence it is an integration test and not a unit-test
func TestSpinnakerPipelineUpdate(t *testing.T) {
	cache.C = cache.NewCache()
	err := dinghyfile.DownloadAndUpdate(&Push{}, &FileService{})
	assert.Equal(t, nil, err)
}
