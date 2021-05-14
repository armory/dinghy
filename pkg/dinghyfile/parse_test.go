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
	"errors"
	"github.com/armory/dinghy/pkg/dinghyfile/pipebuilder"
	"path/filepath"
	"strings"
	"testing"

	"encoding/json"

	"github.com/armory/dinghy/pkg/git/dummy"
	"github.com/armory/plank/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var githubPayloadtest = `
{
  "ref": "refs/tags/simple-tag",
  "before": "6113728f27ae82c7b1a177c8d03f9e96e0adf246",
  "after": "0000000000000000000000000000000000000000",
  "created": false,
  "deleted": true,
  "forced": false,
  "base_ref": null,
  "compare": "https://github.com/Codertocat/Hello-World/compare/6113728f27ae...000000000000",
  "commits": [

  ],
  "head_commit": null,
  "repository": {
    "id": 186853002,
    "node_id": "MDEwOlJlcG9zaXRvcnkxODY4NTMwMDI=",
    "name": "Hello-World",
    "full_name": "Codertocat/Hello-World",
    "private": false,
    "owner": {
      "name": "Codertocat",
      "email": "21031067+Codertocat@users.noreply.github.com",
      "login": "Codertocat",
      "id": 21031067,
      "node_id": "MDQ6VXNlcjIxMDMxMDY3",
      "avatar_url": "https://avatars1.githubusercontent.com/u/21031067?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/Codertocat",
      "html_url": "https://github.com/Codertocat",
      "followers_url": "https://api.github.com/users/Codertocat/followers",
      "following_url": "https://api.github.com/users/Codertocat/following{/other_user}",
      "gists_url": "https://api.github.com/users/Codertocat/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/Codertocat/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/Codertocat/subscriptions",
      "organizations_url": "https://api.github.com/users/Codertocat/orgs",
      "repos_url": "https://api.github.com/users/Codertocat/repos",
      "events_url": "https://api.github.com/users/Codertocat/events{/privacy}",
      "received_events_url": "https://api.github.com/users/Codertocat/received_events",
      "type": "User",
      "site_admin": false
    },
    "html_url": "https://github.com/Codertocat/Hello-World",
    "description": null,
    "fork": false,
    "url": "https://github.com/Codertocat/Hello-World",
    "forks_url": "https://api.github.com/repos/Codertocat/Hello-World/forks",
    "keys_url": "https://api.github.com/repos/Codertocat/Hello-World/keys{/key_id}",
    "collaborators_url": "https://api.github.com/repos/Codertocat/Hello-World/collaborators{/collaborator}",
    "teams_url": "https://api.github.com/repos/Codertocat/Hello-World/teams",
    "hooks_url": "https://api.github.com/repos/Codertocat/Hello-World/hooks",
    "issue_events_url": "https://api.github.com/repos/Codertocat/Hello-World/issues/events{/number}",
    "events_url": "https://api.github.com/repos/Codertocat/Hello-World/events",
    "assignees_url": "https://api.github.com/repos/Codertocat/Hello-World/assignees{/user}",
    "branches_url": "https://api.github.com/repos/Codertocat/Hello-World/branches{/branch}",
    "tags_url": "https://api.github.com/repos/Codertocat/Hello-World/tags",
    "blobs_url": "https://api.github.com/repos/Codertocat/Hello-World/git/blobs{/sha}",
    "git_tags_url": "https://api.github.com/repos/Codertocat/Hello-World/git/tags{/sha}",
    "git_refs_url": "https://api.github.com/repos/Codertocat/Hello-World/git/refs{/sha}",
    "trees_url": "https://api.github.com/repos/Codertocat/Hello-World/git/trees{/sha}",
    "statuses_url": "https://api.github.com/repos/Codertocat/Hello-World/statuses/{sha}",
    "languages_url": "https://api.github.com/repos/Codertocat/Hello-World/languages",
    "stargazers_url": "https://api.github.com/repos/Codertocat/Hello-World/stargazers",
    "contributors_url": "https://api.github.com/repos/Codertocat/Hello-World/contributors",
    "subscribers_url": "https://api.github.com/repos/Codertocat/Hello-World/subscribers",
    "subscription_url": "https://api.github.com/repos/Codertocat/Hello-World/subscription",
    "commits_url": "https://api.github.com/repos/Codertocat/Hello-World/commits{/sha}",
    "git_commits_url": "https://api.github.com/repos/Codertocat/Hello-World/git/commits{/sha}",
    "comments_url": "https://api.github.com/repos/Codertocat/Hello-World/comments{/number}",
    "issue_comment_url": "https://api.github.com/repos/Codertocat/Hello-World/issues/comments{/number}",
    "contents_url": "https://api.github.com/repos/Codertocat/Hello-World/contents/{+path}",
    "compare_url": "https://api.github.com/repos/Codertocat/Hello-World/compare/{base}...{head}",
    "merges_url": "https://api.github.com/repos/Codertocat/Hello-World/merges",
    "archive_url": "https://api.github.com/repos/Codertocat/Hello-World/{archive_format}{/ref}",
    "downloads_url": "https://api.github.com/repos/Codertocat/Hello-World/downloads",
    "issues_url": "https://api.github.com/repos/Codertocat/Hello-World/issues{/number}",
    "pulls_url": "https://api.github.com/repos/Codertocat/Hello-World/pulls{/number}",
    "milestones_url": "https://api.github.com/repos/Codertocat/Hello-World/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/Codertocat/Hello-World/notifications{?since,all,participating}",
    "labels_url": "https://api.github.com/repos/Codertocat/Hello-World/labels{/name}",
    "releases_url": "https://api.github.com/repos/Codertocat/Hello-World/releases{/id}",
    "deployments_url": "https://api.github.com/repos/Codertocat/Hello-World/deployments",
    "created_at": 1557933565,
    "updated_at": "2019-05-15T15:20:41Z",
    "pushed_at": 1557933657,
    "git_url": "git://github.com/Codertocat/Hello-World.git",
    "ssh_url": "git@github.com:Codertocat/Hello-World.git",
    "clone_url": "https://github.com/Codertocat/Hello-World.git",
    "svn_url": "https://github.com/Codertocat/Hello-World",
    "homepage": null,
    "size": 0,
    "stargazers_count": 0,
    "watchers_count": 0,
    "language": "Ruby",
    "has_issues": true,
    "has_projects": true,
    "has_downloads": true,
    "has_wiki": true,
    "has_pages": true,
    "forks_count": 1,
    "mirror_url": null,
    "archived": false,
    "disabled": false,
    "open_issues_count": 2,
    "license": null,
    "forks": 1,
    "open_issues": 2,
    "watchers": 0,
    "default_branch": "master",
    "stargazers": 0,
    "master_branch": "master"
  },
  "pusher": {
    "name": "Codertocat",
    "email": "21031067+Codertocat@users.noreply.github.com"
  },
  "sender": {
    "login": "Codertocat",
    "id": 21031067,
    "node_id": "MDQ6VXNlcjIxMDMxMDY3",
    "avatar_url": "https://avatars1.githubusercontent.com/u/21031067?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/Codertocat",
    "html_url": "https://github.com/Codertocat",
    "followers_url": "https://api.github.com/users/Codertocat/followers",
    "following_url": "https://api.github.com/users/Codertocat/following{/other_user}",
    "gists_url": "https://api.github.com/users/Codertocat/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/Codertocat/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/Codertocat/subscriptions",
    "organizations_url": "https://api.github.com/users/Codertocat/orgs",
    "repos_url": "https://api.github.com/users/Codertocat/repos",
    "events_url": "https://api.github.com/users/Codertocat/events{/privacy}",
    "received_events_url": "https://api.github.com/users/Codertocat/received_events",
    "type": "User",
    "site_admin": false
  }
}
`

var fileService = dummy.FileService{
	"master": {
		"missing_module_test": `{ {{ module "missing" }} }`,
		"extra_data_test": `{
		"application": "my fancy application (author: {{ .RawData.pusher.name }})",
		"pipelines": [
			"stages": [
				{{ $mods := makeSlice "mod1" "mod2" }}
				{{ range $mods }}
					{{ module . }}
				{{ end }}
			]	
			{{ module "deep.pipeline.module" 
				"artifact" "artifact11"
				"artifact2" "artifact22"
			}}
		]
	}`,
		"if_test": `{
		{{ if eq .Branch "master" }}
		"stages": [
			{{ module "mod1" }},
			{{ module "mod2" }}
		]
		{{ end }}
    }`,
		"range_test": `{
		"stages": [
			{{ $mods := makeSlice "mod1" "mod2" }}
			{{ range $mods }}
				{{ module . }}
			{{ end }}
		]
	}`,
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
		"df_local_module": `{
		"application": "search",
		"pipelines": [
			{{ local_module "mod1" }},
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
		//This is for one dinghy file having an if-else conditional and being true
		"if_params_indinghyfiletrue.dinghyfile": `{
		  {{ if eq "test" "test" }}
		  "test": "true"
		  {{ else }}
		  "test": "false"
		  {{ end }}
	}`,
		//This is for one dinghy file having an if-else conditional and being false
		"if_params_indinghyfilefalse.dinghyfile": `{
		  {{ if eq "teste" "test" }}
		  "test": "true"
		  {{ else }}
		  "test": "false"
		  {{ end }}
	}`,
		//Test RawData and a conditional of it with test pusher name
		"rawData.dinghyfile": `{
		  "testprint" : "{{ .RawData.pusher.name }}",
		  {{ if eq .RawData.pusher.name "Codertocat" }}
			"test": "true"
		  {{ else }}
			"test": "false"
		  {{ end }}
	}`,
		//Test no space parsing in dinghyfile
		"no_space.dinghyfile": `{
  "testprint" : "{{.RawData.pusher.name}}"
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

		// Testing the pipelineID function
		"pipelineIDTest": `{
		"application": "pipelineidexample",
		"failPipeline": true,
		"name": "Pipeline",
		"pipeline": "{{ pipelineID "triggerApp" "trigger Pipeline" }}",
		"refId": "1",
		"requisiteStageRefIds": [],
		"type": "pipeline",
		"waitForCompletion": true
	}`,

		// RenderPreprocessFail
		"preprocess_fail": `{
		{{ 
	}`,

		// RenderParseGlobalVarsFail
		"global_vars_parse_fail": `
		["foo", "bar"]
	`,

		// RenderGlobalVarsExtractFail
		"global_vars_extract_fail": `{
		"globals": 42
	}`,

		// VarFuncNotDefined
		"varfunc_not_defined": `{
	  "test": {{ var "biff" }}
	}`,

		// TemplateParseFail
		"template_parse_fail": `{
	  "test": {{ nope "biff" }}
	}`,

		// TemplateBufferFail
		"template_buffer_fail": `{
	  "test": {{ if 4 gt 3 }} "biff" {{ end }}
	}`,

		// OddParamsError
		"odd_params_error": "",

		// DictKeysError
		"dict_keys_error": "",

		// Sprig functions test
		"sprig_functions": `{
			"test": {{ splitList "$" "foo$bar$baz" | toJson }}
		}`,
		// ParseModuleOnValidation
		"parse_module_on_validation": `{
		  "name": value,
		  "waitTime": 1,
		  "type": 1
		}`,
	},
	"branch": {
		"if_test": `{
		{{ if eq .Branch "master" }}
		"stages": [
			{{ module "mod1" }},
			{{ module "mod2" }}
		]
		{{ end }}
    }`,
		"different_branch": `{
		"stages: [
		  {{ module "mod1" }}
		]
		}`,
	},
}

// This returns a test PipelineBuilder object.
func testPipelineBuilder() *PipelineBuilder {
	pb := testBasePipelineBuilder()
	pb.Downloader = fileService
	return pb
}

// For the most part, this is the base object to test against; you may need
// to set things in .Builder from here (see above) after-the-fact.
func testDinghyfileParser() *DinghyfileParser {
	return NewDinghyfileParser(testPipelineBuilder())
}

func TestGracefulErrorHandling(t *testing.T) {
	builder := testDinghyfileParser()
	_, err := builder.Parse("org", "repo", "df_bad", "master", nil)
	assert.NotNil(t, err, "Got non-nil output for mal-formed template action in df_bad")
}

func TestNestedVars(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "nested_var_df"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, _ := r.Parse("org", "repo", "nested_var_df", "master", nil)

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
			r := testDinghyfileParser()
			r.Builder.DinghyfileName = filepath.Base(c.filename)

			buf, _ := r.Parse("org", "repo", c.filename, "master", nil)
			exp := strings.Join(strings.Fields(c.expected), "")
			actual := strings.Join(strings.Fields(buf.String()), "")
			assert.Equal(t, exp, actual)
		})
	}
}

func TestSimpleWaitStage(t *testing.T) {
	r := testDinghyfileParser()
	buf, _ := r.Parse("org", "repo", "df3", "master", nil)

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
	r := testDinghyfileParser()
	buf, _ := r.Parse("org", "repo", "df", "master", nil)

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

func TestMissingModule(t *testing.T) {
	r := testDinghyfileParser()
	buf, err := r.Parse("org", "repo", "missing_module_test", "master", nil)
	assert.Error(t, err)
	assert.Nil(t, buf)
}

func TestUnconfiguredTemplateOrg(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.TemplateOrg = ""
	buf, err := r.Parse("org", "repo", "missing_module_test", "master", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot load module missing; templateOrg not configured")
	assert.Nil(t, buf)
}

func TestIfConditionEmpty(t *testing.T) {
	r := testDinghyfileParser()

	// if we don't match the if condition this should be blank
	buf, err := r.Parse("org", "repo", "if_test", "branch", nil)
	require.Nil(t, err)
	expected := `{}`
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)

	// if we do match the if condition, we should see some stage data
	buf, err = r.Parse("foo", "repo", "if_test", "master", nil)
	require.Nil(t, err)
	expected = `{"stages":[{"foo":"bar","type":"deploy"},{"type":"jenkins"}]}`
	exp = strings.Join(strings.Fields(expected), "")
	actual = strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

func TestConditionalExtraDataFromGit(t *testing.T) {
	r := testDinghyfileParser()
	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(githubPayloadtest), &d)
	assert.Nil(t, err)
	r.Builder.PushRaw = d

	buf, _ := r.Parse("org", "repo", "extra_data_test", "master", nil)
	expected := `{"application":"myfancyapplication(author:Codertocat)","pipelines":["stages":[{"foo":"bar","type":"deploy"}{"type":"jenkins"}]{"parameterConfig":[{"description":"ServiceName","name":"service","required":true,{"parameterConfig":[{"artifact":artifact11,}}",{"parameterConfig":[{"artifact":artifact22,}}",}}]}`
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

func TestRangeSyntax(t *testing.T) {
	r := testDinghyfileParser()
	buf, err := r.Parse("org", "repo", "range_test", "master", nil)
	require.Nil(t, err)

	const expected = `{"stages":[{"foo":"bar","type":"deploy"}{"type":"jenkins"}]}`
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
	r := testDinghyfileParser()
	ts := testStruct{}
	ret, err := r.Parse("org", "repo", "df2", "master", nil)
	err = json.Unmarshal(ret.Bytes(), &ts)
	assert.Equal(t, nil, err)

	assert.Equal(t, "baz", ts.Foo)
	assert.Equal(t, "b", ts.A)
	assert.Equal(t, 100, ts.Nested.WaitTime)
}

func TestPipelineIDFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetPipelines(gomock.Eq("triggerApp"), "").Return([]plank.Pipeline{plank.Pipeline{ID: "pipelineID", Name: "trigger Pipeline"}}, nil).Times(1)
	r.Builder.Client = client

	vars := []VarMap{
		{"triggerApp": "triggerApp", "trigger Pipeline": "trigger Pipeline"},
	}
	idFunc := r.pipelineIDFunc(vars).(func(string, string) string)
	result := idFunc("triggerApp", "trigger Pipeline")
	assert.Equal(t, "pipelineID", result)
}
func TestPipelineIDFuncDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetPipelines(gomock.Eq("triggerApp"), "").Return(nil, errors.New("fake not found")).Times(1)
	r.Builder.Client = client

	vars := []VarMap{
		{"triggerApp": "triggerApp"}, {"trigger Pipeline": "trigger Pipeline"},
	}
	idFunc := r.pipelineIDFunc(vars).(func(string, string) string)
	result := idFunc("triggerApp", "trigger Pipeline")
	assert.Equal(t, "", result)
}
func TestPipelineIDRender(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetPipelines(gomock.Eq("triggerApp"), "").Return([]plank.Pipeline{plank.Pipeline{ID: "pipeline ID", Name: "trigger Pipeline"}}, nil).Times(1)
	r.Builder.Client = client

	expected := `{
		"application": "pipelineidexample",
		"failPipeline": true,
		"name": "Pipeline",
		"pipeline": "pipeline ID",
		"refId": "1",
		"requisiteStageRefIds": [],
		"type": "pipeline",
		"waitForCompletion": true
	}`

	ret, err := r.Parse("org", "repo", "pipelineIDTest", "master", nil)
	assert.Nil(t, err)
	assert.Equal(t, expected, ret.String())
}

func TestModuleEmptyString(t *testing.T) {
	r := testDinghyfileParser()
	ret, _ := r.Parse("org", "repo", "df4", "master", nil)
	assert.Equal(t, `{"foo": ""}`, ret.String())
}

func TestLocalModule(t *testing.T) {
	r := testDinghyfileParser()
	expected := `{
		"application": "search",
		"pipelines": [
			{
		"foo": "bar",
		"type": "deploy"
	},
			{
		"type": "foobar"
	}
		]
	}`
	ret, _ := r.Parse("org", "repo", "df_local_module", "master", nil)
	assert.Equal(t, expected, ret.String())
}

func TestDeepVars(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "deep_var_df"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, _ := r.Parse("org", "repo", "deep_var_df", "master", nil)

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
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "deep_var_df"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, _ := r.Parse("org", "repo", "empty_default_variables", "master", nil)

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
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "if_params.dinghyfile"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, err := r.Parse("org", "repo", "if_params.dinghyfile", "master", nil)
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

func TestConditionalArgsInDinghyfileTrue(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "if_params_indinghyfiletrue.dinghyfile"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, err := r.Parse("org", "repo", "if_params_indinghyfiletrue.dinghyfile", "master", nil)
	require.Nil(t, err)

	const raw = `{
  "test": "true"
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

func TestConditionalArgsInDinghyfileFalse(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "if_params_indinghyfilefalse.dinghyfile"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, err := r.Parse("org", "repo", "if_params_indinghyfilefalse.dinghyfile", "master", nil)
	require.Nil(t, err)

	const raw = `{
  "test": "false"
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

func TestNoSpaceBeforeCurlyBrackets(t *testing.T) {
	// deserialze push data to a map.  used in template logic later
	rawPushData := make(map[string]interface{})
	err := json.Unmarshal([]byte(githubPayloadtest), &rawPushData)
	require.Nil(t, err)

	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "no_space.dinghyfile"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	r.Builder.PushRaw = rawPushData
	buf, err := r.Parse("org", "repo", "no_space.dinghyfile", "master", nil)
	require.Nil(t, err)

	const raw = `{
  "testprint" : "Codertocat"
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

func TestRawDataInDinghyfile(t *testing.T) {
	// deserialze push data to a map.  used in template logic later
	rawPushData := make(map[string]interface{})
	err := json.Unmarshal([]byte(githubPayloadtest), &rawPushData)
	require.Nil(t, err)

	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "rawData.dinghyfile"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	r.Builder.PushRaw = rawPushData
	buf, err := r.Parse("org", "repo", "rawData.dinghyfile", "master", nil)
	require.Nil(t, err)

	const raw = `{
  "testprint" : "Codertocat",
  "test": "true"
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "var_params.outer"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"

	// Despite this "error", no warnings or errors are expected to show up.
	// TODO:  We should perhaps figure out how to identify this failure case?
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Warnf(gomock.Any()).Times(0)
	logger.EXPECT().Errorf(gomock.Eq("Error parsing value %s: %s"), gomock.Any(), gomock.Any()).Times(1)
	logger.EXPECT().Error(gomock.Any()).Times(0)
	logger.EXPECT().Info(gomock.Eq("No global vars found in dinghyfile")).Times(1)

	buf, err := r.Parse("org", "repo", "var_params.outer", "master", nil)
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

func TestRenderPreprocessFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "preprocess_fail"
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Failed to preprocess:\n %s"), gomock.Any()).Times(1)

	_, err := r.Parse("org", "repo", "preprocess_fail", "master", nil)
	assert.NotNil(t, err)
}

func TestRenderParseGlobalVarsFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "global_vars_parse_fail"
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Failed to parse global vars:\n %s"), gomock.Any()).Times(1)

	_, err := r.Parse("org", "repo", "global_vars_parse_fail", "master", nil)
	assert.NotNil(t, err)
}

func TestRenderGlobalVarsExtractFail(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "global_vars_extract_fail"

	_, err := r.Parse("org", "repo", "global_vars_extract_fail", "master", nil)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "could not extract global vars from:\n {\n\t\t\"globals\": 42\n\t}")
}

func TestRenderVarFuncNotDefined(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "varfunc_not_defined"

	buf, err := r.Parse("org", "repo", "varfunc_not_defined", "master", nil)
	require.Nil(t, err)

	var actual interface{}
	err = json.Unmarshal(buf.Bytes(), &actual)
	// This errors because the resulting JSON is { "test": } (since the var
	// gets replaced with nothing at all) and this is invalid.
	require.NotNil(t, err)
}

func TestRenderDownloadFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Failed to download %s/%s/%s/%s"), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	_, err := r.Parse("org", "repo", "nonexistentfile", "master", nil)
	require.NotNil(t, err)
	require.Equal(t, "File not found", err.Error())
}

func TestRenderTemplateParseFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Failed to parse template:\n %s"), gomock.Any()).Times(1)

	_, err := r.Parse("org", "repo", "template_parse_fail", "master", nil)
	require.NotNil(t, err)
	require.Equal(t, "template: dinghy-render:2: function \"nope\" not defined", err.Error())
}

func TestRenderTemplateBufferFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Errorf(gomock.Eq("Failed to execute buffer:\n %s\nError: %s"), gomock.Any(), gomock.Any()).Times(1)

	_, err := r.Parse("org", "repo", "template_buffer_fail", "master", nil)
	require.NotNil(t, err)
	require.Equal(t, "template: dinghy-render:2:17: executing \"dinghy-render\" at <4>: can't give argument to non-function 4", err.Error())
}

func TestRenderValueArrayFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	biff := make(chan int)
	ex := []interface{}{biff}

	r := testDinghyfileParser()
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Errorf(gomock.Eq("unable to json.marshal array value %v"), gomock.Eq(ex)).Times(1)

	res := r.renderValue(ex)
	assert.Equal(t, "", res.(string))
}

func TestRenderValueMapFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	biff := make(chan int)
	ex := map[string]interface{}{"foo": biff}

	r := testDinghyfileParser()
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Errorf(gomock.Eq("unable to json.marshal map value %v"), gomock.Eq(ex)).Times(1)

	res := r.renderValue(ex)
	assert.Equal(t, "", res.(string))
}

func TestModuleFuncOddParamsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_key := "odd_params_error"
	r := testDinghyfileParser()
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Warnf(gomock.Eq("odd number of parameters received to module %s"), gomock.Eq(test_key)).Times(1)

	modFunc := r.moduleFunc("org", "repo", "master", map[string]bool{}, []VarMap{})
	res, _ := modFunc.(func(string, ...interface{}) (string, error))(test_key, "biff")
	assert.Equal(t, "", res)
}

func TestModuleFuncDictKeysError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	test_key := "dict_keys_error"
	r := testDinghyfileParser()
	logger := mockLogger(r, ctrl)
	logger.EXPECT().Errorf(gomock.Eq("dict keys must be strings in module: %s"), gomock.Eq(test_key)).Times(1)

	modFunc := r.moduleFunc("org", "repo", "master", map[string]bool{}, []VarMap{})
	res, _ := modFunc.(func(string, ...interface{}) (string, error))(test_key, 42, "foo")
	assert.Equal(t, "", res)
}

func TestDifferentTemplateBranch(t *testing.T) {
	r := testDinghyfileParser()

	//  This pulls in "mod1" but "mod1" is only on "master", not on "branch".
	_, err := r.Parse("org", "repo", "different_branch", "branch", nil)

	expected := `template: dinghy-render:3:7: executing "dinghy-render" at <module "mod1">: error calling module: error rendering imported module 'mod1': File not found`
	assert.Equal(t, expected, err.Error())

}

func TestSprigFuncs(t *testing.T) {
	r := testDinghyfileParser()
	filepath := "sprig_functions"
	r.Builder.DinghyfileName = filepath
	buf, _ := r.Parse("org", "repo", filepath, "master", nil)

	const expected = `{
		"test": ["foo","bar","baz"]
	}`

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

func TestParseModuleOnValidation(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "nested_var_df"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	r.Builder.Action = pipebuilder.Validate
	r.Builder.TemplateRepo = "repo"
	_, err := r.Parse("org", "repo", "parse_module_on_validation", "master", nil)

	assert.NotNil(t, err)
}
