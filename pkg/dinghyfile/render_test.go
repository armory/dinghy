/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package dinghyfile

import (
	"path/filepath"
	"strings"
	"testing"

	"encoding/json"

	"github.com/armory/dinghy/pkg/cache"
	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/git/dummy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fileService = dummy.FileService{
	"df": `{
		"stages": [
			{{ module "mod1" }},
			{{ module "mod2" }}
		]
	}`,
	"df2": `{{ module "mod4" "foo" "baz" "waitTime" 100 }}`,
	"df3": `{
		"stages": [
			{{ module "mod6" "waitTime" 10 "refId" { "c": "d" } "requisiteStageRefIds" ["1", "2", "3"] }}
		]
	}`,
	"df4": `{{ module "mod3" "foo" "" }}`,
	"df_bad": `{
		"stages": [
			{{ module "mod1" }
		]
	}`,
	"df_global": `{
		"application": "search",
		"globals": {
			"type": "foo"
		},
		"pipelines": [
			{{ module "mod1" }},
			{{ module "mod2" "type" "foobar" }}
			]
	}`,
	"df_spec": `{
		"spec": {
			"name": "search",
			"email": "unknown@unknown.com",
			"dataSources": {
				"disabled":[],
				"enabled":["canaryConfigs"]
			}
		},
		"globals": {
			"type": "foo"
		},
		"pipelines": [
			{{ module "mod1" }},
			{{ module "mod2" "type" "foobar" }}
			]
	}`,
	"df_app_global": `{
		"application": "search",
		{{ appModule "appmod" }},
		"globals": {
			"type": "foo"
		},
		"pipelines": [
			{{ module "mod1" }},
			{{ module "mod2" "type" "foobar" }}
			]
	}`,
	"df_global/nested": `{
		"application": "search",
		"globals": {
			"type": "foo"
		},
		"pipelines": [
			{{ module "mod1" }},
			{{ module "mod2" "type" "foobar" }}
			]
	}`,
	"appmod": `"description": "description"`,
	"mod1": `{
		"foo": "bar",
		"type": "{{ var "type" ?: "deploy" }}"
	}`,
	"mod2": `{
		"type": "{{ var "type" ?: "jenkins" }}"
	}`,
	"mod3": `{"foo": "{{ var "foo" ?: "baz" }}"}`,

	"mod4": `{
		"foo": "{{ var "foo" "baz" }}",
		"a": "{{ var "nonexistent" "b" }}",
		"nested": {{ module "mod5" }}
	}`,

	"mod5": `{
		"waitTime": {{ var "waitTime" 1000 }}
	}`,

	"mod6": `{
		"name": "Wait",
		"refId": {{ var "refId" {} }},
		"requisiteStageRefIds": {{ var "requisiteStageRefIds" [] }},
		"type": "wait",
		"waitTime": {{ var "waitTime" 12044 }}
	}`,

	"nested_var_df": `{
		"application": "dinernotifications",
		"globals": {
			 "application": "dinernotifications"
		 },
		"pipelines": [
			{{ module "preprod_teardown.pipeline.module" }}
		]
	}`,

	"preprod_teardown.pipeline.module": `{
		"parameterConfig": [
			{
				"default": "{{ var "discovery-service-name" ?: "@application" }}",
				"description": "Service Name",
				"name": "service",
				"required": true
			}
		}`,

	"deep_var_df": `{
		"application": "dinernotifications",
		"globals": {
			 "application": "dinernotifications"
		 },
		"pipelines": [
			{{ module "deep.pipeline.module" 
				"artifact" "artifact11"
				"artifact2" "artifact22"
			}}
		]
	}`,

	"deep.pipeline.module": `{
		"parameterConfig": [
			{
				"description": "Service Name",
				"name": "service",
				"required": true,
				{{ module "deep.stage.module" 
					"artifact" {{var artifact}}
				}}",
				{{ module "deep.stage.module" 
					"artifact" {{var artifact2}}
				}}",
			}
		}`,

	"deep.stage.module": `{
		"parameterConfig": [
			{
				"artifact": {{ var "artifact" }},
			}
		}`,

	"empty_default_variables": `{
		"application": "dinernotifications",
		"pipelines": [
			{{ module "empty_default_variables.pipeline.module" }}
		]
	}`,

	"empty_default_variables.pipeline.module": `{
		"parameterConfig": [
			{
				"default": "{{ var "discovery-service-name" ?: "" }}",
				"description": "Service Name",
				"name": "service",
				"required": true
			}
		}`,

	// if_params reproduced/resolved an issue Giphy had where they were trying to use an
	// if conditional inside a {{ module }} call.
	"if_params.dinghyfile": `{
		"test": "if_params",
		"result": {{ module "if_params.midmodule"
								 "straightvar" "foo"
								 "condvar" true }}
	}`,
	// NOTE:  This next example is a _functional_ way to do conditional arguments to a module.
	// This is the result of trying to debug why this markup didn't work properly:
	//    {{ module "if_params.bottom"
	//              "foo" "bar"
	//              {{ if var "condvar" }}
	//              "extra" ["foo", "bar"]
	//              {{ end }}
	//   }}
	// The reason is that nested template markup isn't evaluated inside-out, so the argument
	// to "module" is actually the string "{{ if var "condvar" }}"
	"if_params.midmodule": `
		{{ if var "condvar" }}
		{{ module "if_params.bottom"
								 "foo" "bar"
								 "extra" [ "foo", "bar" ]
		}}
		{{ else }}
		{{ module "if_params.bottom" "foo" "bar" }}
		{{ end }}
	`,
	"if_params.bottom": `{
		"foo": "{{ var "foo" ?: "default" }}",
		"biff": {{ var "extra" ?: ["NotSet"] }}
	}`,

	// var_params tests whether or not you can reference a variable inside a value
	// being sent to a module.  The answer is "no"; and this will result in invalid JSON.
	"var_params.outer": `
		{{ module "var_params.middle" "myvar" "success" }}
	`,
	"var_params.middle": `
		{{ module "var_params.inner"
							"foo" [ { "bar": {{ var "myvar" ?: "failure"}} } ]
		}}
	`,
	"var_params.inner": `{
		"foo": {{ var "foo" }}
	}`,
}

// mock out events so that it gets passed over and doesn't do anything
type EventsTestClient struct{}

func (c *EventsTestClient) SendEvent(eventType string, event *events.Event) {}

func TestGracefulErrorHandling(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:     cache.NewMemoryCache(),
		Downloader: fileService,
	}
	_, err := builder.Render("org", "repo", "df_bad", nil)
	assert.NotNil(t, err, "Got non-nil output for mal-formed template action in df_bad")
}

func TestNestedVars(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:         cache.NewMemoryCache(),
		Downloader:     fileService,
		DinghyfileName: "nested_var_df",
		TemplateOrg:    "org",
		TemplateRepo:   "repo",
		EventClient:    &EventsTestClient{},
	}
	buf, _ := builder.Render("org", "repo", "nested_var_df", nil)

	const expected = `{
		"application": "dinernotifications",
		"globals": {
			 "application": "dinernotifications"
		 },
		"pipelines": [
			{
				"parameterConfig": [
					{
						"default": "dinernotifications",
						"description": "Service Name",
						"name": "service",
						"required": true
					}
			}
		]
	}`

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

func TestGlobalVars(t *testing.T) {

	cases := map[string]struct {
		filename string
		expected string
	}{
		"df_global": {
			filename: "df_global",
			expected: `{
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
				}`,
		},
		"df_global_nested": {
			filename: "df_global/nested",
			expected: `{
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
				}`,
		},
		"df_global_appmodule": {
			filename: "df_app_global",
			expected: `{
				"application": "search",
				"description": "description",
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
				}`,
		},
		"df_spec": {
			filename: "df_spec",
			expected: `{
				"spec": {
					"name": "search",
					"email": "unknown@unknown.com",
					"dataSources": {
						"disabled":[],
						"enabled":["canaryConfigs"]
					}
				},
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
				}`,
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			builder := &PipelineBuilder{
				Depman:         cache.NewMemoryCache(),
				Downloader:     fileService,
				DinghyfileName: filepath.Base(c.filename),
				EventClient:    &EventsTestClient{},
			}

			buf, _ := builder.Render("org", "repo", c.filename, nil)
			exp := strings.Join(strings.Fields(c.expected), "")
			actual := strings.Join(strings.Fields(buf.String()), "")
			assert.Equal(t, exp, actual)
		})
	}
}

func TestSimpleWaitStage(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:      cache.NewMemoryCache(),
		Downloader:  fileService,
		EventClient: &EventsTestClient{},
	}
	buf, _ := builder.Render("org", "repo", "df3", nil)

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
		Depman:      cache.NewMemoryCache(),
		Downloader:  fileService,
		EventClient: &EventsTestClient{},
	}
	buf, _ := builder.Render("org", "repo", "df", nil)

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

type testStruct struct {
	Foo    string `json:"foo"`
	A      string `json:"a"`
	Nested struct {
		WaitTime int `json:"waitTime"`
	} `json:"nested"`
}

func TestModuleVariableSubstitution(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:      cache.NewMemoryCache(),
		Downloader:  fileService,
		EventClient: &EventsTestClient{},
	}
	ts := testStruct{}
	ret, err := builder.Render("org", "repo", "df2", nil)
	err = json.Unmarshal(ret.Bytes(), &ts)
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

func TestModuleEmptyString(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:      cache.NewMemoryCache(),
		Downloader:  fileService,
		EventClient: &EventsTestClient{},
	}
	ret, _ := builder.Render("org", "repo", "df4", nil)
	assert.Equal(t, `{"foo": ""}`, ret.String())
}

func TestDeepVars(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:         cache.NewMemoryCache(),
		Downloader:     fileService,
		DinghyfileName: "deep_var_df",
		TemplateOrg:    "org",
		TemplateRepo:   "repo",
		EventClient:    &EventsTestClient{},
	}
	buf, _ := builder.Render("org", "repo", "deep_var_df", nil)

	const expected = `{
		"application": "dinernotifications",
		"globals": {
			 "application": "dinernotifications"
		 },
		"pipelines": [
			{
		"parameterConfig": [
			{
				"description": "Service Name",
				"name": "service",
				"required": true,
				{
		"parameterConfig": [
			{
				"artifact": artifact11,
			}
		}",
				{
		"parameterConfig": [
			{
				"artifact": artifact22,
			}
		}",
			}
		}
		]
	}`

	// strip whitespace from both strings for assertion
	exp := expected
	actual := buf.String()
	assert.Equal(t, exp, actual)
}

func TestEmptyDefaultVar(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:         cache.NewMemoryCache(),
		Downloader:     fileService,
		DinghyfileName: "deep_var_df",
		TemplateOrg:    "org",
		TemplateRepo:   "repo",
		EventClient:    &EventsTestClient{},
	}
	buf, _ := builder.Render("org", "repo", "empty_default_variables", nil)

	const expected = `{
		"application": "dinernotifications",
		"pipelines": [
			{
		"parameterConfig": [
			{
				"default": "",
				"description": "Service Name",
				"name": "service",
				"required": true
			}
		}
		]
	}`

	exp := expected
	actual := buf.String()
	assert.Equal(t, exp, actual)
}

func TestConditionalArgs(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:         cache.NewMemoryCache(),
		Downloader:     fileService,
		DinghyfileName: "if_params.dinghyfile",
		TemplateOrg:    "org",
		TemplateRepo:   "repo",
		EventClient:    &EventsTestClient{},
	}
	buf, err := builder.Render("org", "repo", "if_params.dinghyfile", nil)
	require.Nil(t, err)

	const raw = `{
		 "test": "if_params",
		 "result": {
			 "foo": "bar",
			 "biff": ["foo", "bar"]
		 }
	}`
	var expected interface{}
	err = json.Unmarshal([]byte(raw), &expected)
	require.Nil(t, err)
	expected_str, err := json.Marshal(expected)
	require.Nil(t, err)

	var actual interface{}
	err = json.Unmarshal(buf.Bytes(), &actual)
	require.Nil(t, err)
	actual_str, err := json.Marshal(actual)
	require.Nil(t, err)

	require.Equal(t, string(expected_str), string(actual_str))
}

// TODO:  This test is currently a negative test -- the example inputs do NOT work properly,
//        and currently, this is expected behavior; we should change this test when we decide
//        if a) we should be catching the error in the Render, or b) we should handle this
//        kind of nested markup.
func TestVarParams(t *testing.T) {
	builder := &PipelineBuilder{
		Depman:         cache.NewMemoryCache(),
		Downloader:     fileService,
		DinghyfileName: "var_params.outer",
		TemplateOrg:    "org",
		TemplateRepo:   "repo",
		EventClient:    &EventsTestClient{},
	}
	buf, err := builder.Render("org", "repo", "var_params.outer", nil)
	// Unfortunately, we don't currently catch this failure here.
	assert.Nil(t, err)

	var actual interface{}
	err = json.Unmarshal(buf.Bytes(), &actual)
	assert.NotNil(t, err)
	/* TODO:  Uncomment this section when/if we make nested references work, delete this if
						we test for the error properly.
	require.Nil(t, err)
	actual_str, err := json.Marshal(actual)
	require.Nil(t, err)

	const raw = `{
		"test": [ { "bar": "success" } ]
	}`
	var expected interface{}
	err = json.Unmarshal([]byte(raw), &expected)
	require.Nil(t, err)
	expected_str, err := json.Marshal(expected)
	require.Nil(t, err)

	require.Equal(t, string(expected_str), string(actual_str))
	*/
}
