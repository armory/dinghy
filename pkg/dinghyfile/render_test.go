package dinghyfile

import (
	"strings"
	"testing"

	"encoding/json"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/git/dummy"
	"github.com/stretchr/testify/assert"
	"github.com/armory-io/dinghy/pkg/settings"
)

var fileService = dummy.FileService{
	"simpleTempl": `{
		"stages": [
			{{ module "wait.stage.module" "waitTime" 10 "refId" { "c": "d" } "requisiteStageRefIds" ["1", "2", "3"] }}
		]
	}`,
	"df": `{
		"stages": [
			{{ module "mod1" }},
			{{ module "mod2" }}
		]
	}`,
	"mod1": `{
		"foo": "bar",
		"type": "{{ var "type" ?: "deploy" }}"
	}`,
	"mod2": `{
		"type": "{{ var "type" ?: "jenkins" }}"
	}`,
	"wait.stage.module": `{
		"name": "Wait",
		"refId": {{ var "refId" {} }},
		"requisiteStageRefIds": {{ var "requisiteStageRefIds" [] }},
		"type": "wait",
		"waitTime": {{ var "waitTime" 12044 }}
	}`,
	"gv_dinghy": `{
		"application": "search",
		"globals": {
			"type": "foo"
		},
		"pipelines": [
			{{ module "mod1" }},
			{{ module "mod2" "type" "foobar" }}
	    ]
	  }`,
}


func TestGlobalVars(t *testing.T) {
	builder := &PipelineBuilder{
		Downloader: fileService,
		Depman:     cache.NewMemoryCache(),
	}

	settings.S.DinghyFilename = "gv_dinghy"
	buf := builder.Render("org", "repo", "gv_dinghy", nil)

	const expected = `{
		"application": "search",
		"globals": {
			"type": "foo"
		},
		"pipelines": [
			{
				"foo": "bar",
				"type": "foo"
			},
			{
				"type": "foobar"
			}
	    ]
	  }`

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}


func TestSimpleWaitStage(t *testing.T) {
	builder := &PipelineBuilder{
		Downloader: fileService,
		Depman:     cache.NewMemoryCache(),
	}

	buf := builder.Render("org", "repo", "simpleTempl", nil)

	const expected = `{
		"stages": [
			{
				"name": "Wait",
				"refId": { "c": "d" },
				"requisiteStageRefIds": ["1", "2", "3"],
				"type": "wait",
				"waitTime": 10
			}
		]
	}`

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

func TestSpillover(t *testing.T) {
	builder := &PipelineBuilder{
		Downloader: fileService,
		Depman:     cache.NewMemoryCache(),
	}

	buf := builder.Render("org", "repo", "df", nil)

	const expected = `{
		"stages": [
			{"foo":"bar","type":"deploy"},
			{"type":"jenkins"}
		]
	}`

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

var multilevelFileService = dummy.FileService{
	"dinghyfile": `{{ module "wait.stage.module" "foo" "baz" "waitTime" 100 }}`,

	"wait.stage.module": `{
		"foo": "{{ var "foo" "baz" }}",
		"a": "{{ var "nonexistent" "b" }}",
		"nested": {{ module "wait.dep.module" }}
	}`,

	"wait.dep.module": `{
		"waitTime": {{ var "waitTime" 1000 }}
	}`,
}

type testStruct struct {
	Foo    string `json:"foo"`
	A      string `json:"a"`
	Nested struct {
		WaitTime int `json:"waitTime"`
	} `json:"nested"`
}

func TestModuleVariableSubstitution(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:     cache.NewMemoryCache(),
		Downloader: multilevelFileService,
	}

	ts := testStruct{}
	ret := builder.Render("org", "repo", "dinghyfile", nil)
	err := json.Unmarshal(ret.Bytes(), &ts)
	assert.Equal(t, nil, err)

	assert.Equal(t, "baz", ts.Foo)
	assert.Equal(t, "b", ts.A)
	assert.Equal(t, 100, ts.Nested.WaitTime)
}

/*
func TestPipelineID(t *testing.T) {
	id := pipelineIDFunc("armoryspinnaker", "fake-echo-test")
	assert.Equal(t, "f9c05bd0-5a50-4540-9e15-b44740abfb10", id)
}
*/

var emptyStringFileService = dummy.FileService{
	"dinghyfile":        `{{ module "wait.stage.module" "foo" "" }}`,
	"wait.stage.module": `{"foo": "{{ var "foo" ?: "baz" }}"}`,
}

func TestModuleEmptyString(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:     cache.NewMemoryCache(),
		Downloader: emptyStringFileService,
	}

	ret := builder.Render("org", "repo", "dinghyfile", nil)
	assert.Equal(t, `{"foo": ""}`, ret.String())
}
