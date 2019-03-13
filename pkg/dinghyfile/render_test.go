package dinghyfile

import (
	"path/filepath"
	"strings"
	"testing"

	"encoding/json"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/git/dummy"
	"github.com/stretchr/testify/assert"
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
}

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
		Depman:     cache.NewMemoryCache(),
		Downloader: fileService,
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
		Depman:     cache.NewMemoryCache(),
		Downloader: fileService,
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
		Depman:     cache.NewMemoryCache(),
		Downloader: fileService,
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
		Depman:     cache.NewMemoryCache(),
		Downloader: fileService,
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
	}
	buf, _ := builder.Render("org", "repo", "deep_var_df", nil)

	const expected = `{"application": "dinernotifications", 
					"globals": {
						"application":"dinernotifications"
					},
					"pipelines": [
						{
							"parameterConfig": [
								{
									"description": "ServiceName",
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
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}
