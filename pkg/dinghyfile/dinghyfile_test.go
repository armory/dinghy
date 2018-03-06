package dinghyfile

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const simpleTempl = `
{
    "stages": [
        {{ module "../../example/wait.stage.module" "waitTime" 10 }}
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

func TestSimpleWaitStage(t *testing.T) {
	buf := Render(simpleTempl, "", "", &FileService{})

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}
