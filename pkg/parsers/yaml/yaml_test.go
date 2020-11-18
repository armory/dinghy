package yaml

import (
	"errors"
	"github.com/armory/dinghy/pkg/cache"
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/armory/dinghy/pkg/git/dummy"
	"github.com/armory/dinghy/pkg/preprocessor"
	"github.com/armory/plank/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"strings"
	"testing"
)

var fileService = dummy.FileService{
	"branch": {
		"df": `
stages:
- {{ module "mod1" }}
- {{ module "mod2" }}`,
		"df2": `{{ module "mod4" "foo" "baz" "waitTime" 100 }}`,
		"df3": `
stages:
- {{ module "mod6" "waitTime" 10 "refId" "c:d" "requisiteStageRefIds" ["1", "2", "3"] }}
    `,
		"df4": `{{ module "mod3" "foo" "" }}`,
		"df_bad": `{
		"stages": [
			{{ module "mod1" }
		]
	}`,
		"df_global": `
application: search
globals:
  type: foo
pipelines:
- {{ module "mod1" }}
- {{ module "mod2" "type" "foobar" }}`,

		"df_spec": `
spec:
  name: search
  email: unknown@unknown.com
  dataSources: 
    disabled: []
    enabled:
    - canaryConfigs
globals:
  type: foo
pipelines:
- {{ module "mod1" }}
- {{ module "mod2" "type" "foobar" }}
    `,
		"df_app_global": `
application: search
{{ appModule "appmod" }}
globals:
  type: foo
pipelines:
- {{ module "mod1" }}
- {{ module "mod2" "type" "foobar" }}`,
		"df_global/nested": `
application: search
globals:
  type: foo
pipelines:
- {{ module "mod1" }}
- {{ module "mod2" "type" "foobar" }}
`,
		"appmod": `description: description`,
		"mod1": `
		foo: bar
		type: {{ var "type" ?: "deploy" }}`,
		"mod2": `type: {{ var "type" ?: "jenkins" }}`,
		"mod3": `foo: "{{ var "foo" ?: "baz" }}"`,

		"mod4": `{
		"foo": "{{ var "foo" "baz" }}",
		"a": "{{ var "nonexistent" "b" }}",
		"nested": {{ module "mod5" }}
	}`,

		"mod5": `{
		"waitTime": {{ var "waitTime" 1000 }}
	}`,

		"mod6": `
name: Wait
refId: {{ var "refId" {} }}
requisiteStageRefIds: {{ var "requisiteStageRefIds" [] }}
type: wait
waitTime: {{ var "waitTime" 12044 }}
`,

		"nested_var_df": `
---
application: dinernotifications
globals:
  application: dinernotifications
pipelines:
- {{ module "preprod_teardown.pipeline.module" }}`,

		"preprod_teardown.pipeline.module": `
parameterConfig:
- default: {{ var "discovery-service-name" ?: "@application" }}
  description: Service Name
  name: service
  required: true`,

		"deep_var_df": `
application: dinernotifications
globals:
  application: dinernotifications
pipelines:
{{ module "deep.pipeline.module" "artifact" "artifact11" "artifact2" "artifact22" }}`,

		"deep.pipeline.module": `
- parameterConfig:
  - description: Service Name
    name: service
    required: true
    {{ module "deep.stage.module" "artifact" {{var artifact}} }}
    {{ module "deep.stage.module" "artifact" {{var artifact2}} }}
`,

		"deep.stage.module": `
- parameterConfig:
  - artifact: {{ var "artifact" }}
`,

		"empty_default_variables": `
application: dinernotifications
pipelines:
{{ module "empty_default_variables.pipeline.module" }}
`,

		"empty_default_variables.pipeline.module": `
- parameterConfig:
  - default: "{{ var "discovery-service-name" ?: "" }}"
    description: Service Name
    name: service
    required: true
`,

		// if_params reproduced/resolved an issue Giphy had where they were trying to use an
		// if conditional inside a {{ module }} call.
		"if_params.dinghyfile": `
test: if_params
result:
{{ module "if_params.midmodule"
						 "straightvar" "foo"
						 "condvar" true }}`,
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
{{ module "if_params.bottom" "foo" "bar" "extra" [ "foo", "bar" ] }}
{{ else }}
{{ module "if_params.bottom" "foo" "bar" }}
{{ end }}`,
		"if_params.bottom": `
- foo: {{ var "foo" ?: "default" }}
  biff: {{ var "extra" ?: ["NotSet"] }}
`,

		// var_params tests whether or not you can reference a variable inside a value
		// being sent to a module.  The answer is "no"; and this will result in invalid JSON.
		"var_params.outer": `{{ module "var_params.middle" "myvar" "success" }}`,
		"var_params.middle": `
		{{ module "var_params.inner"
							"foo" [ bar: {{ var "myvar" ?: "failure"}} ]
		}}
	`,
		"var_params.inner": `foo: {{ var "foo" }}`,

		// Testing the pipelineID function
		"pipelineIDTest": `
application: pipelineidexample
failPipeline: true
name: Pipeline
pipeline: {{ pipelineID "triggerApp" "triggerPipeline" }}
refId: 1
requisiteStageRefIds: []
type: pipeline
waitForCompletion: true`,

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

		// Inserting entire pipeline
		"pipeline_insert_base": `
application: "foo"
pipelines:
- application: "foo"
  name: "bar"
  stages:
{{ module "pipeline_insert_mod" }}
  triggers: []
`,

		"pipeline_insert_mod": `  - name: "stage"
    type: wait`,
	},
}

// This returns a test PipelineBuilder object.
func testBasePipelineBuilder() *dinghyfile.PipelineBuilder {
	return &dinghyfile.PipelineBuilder{
		Depman:      cache.NewMemoryCache(),
		EventClient: &dinghyfile.EventsTestClient{},
		Logger:      dinghyfile.NewDinghylog(),
		Ums:         []dinghyfile.Unmarshaller{&DinghyYaml{}},
	}
}

// This returns a test PipelineBuilder object.
func testPipelineBuilder() *dinghyfile.PipelineBuilder {
	pb := testBasePipelineBuilder()
	pb.Downloader = fileService
	return pb
}

// For the most part, this is the base object to test against; you may need
// to set things in .Builder from here (see above) after-the-fact.
func testDinghyfileParser() *DinghyfileYamlParser {
	return NewDinghyfileYamlParser(testPipelineBuilder())
}

func TestGracefulErrorHandling(t *testing.T) {
	builder := testDinghyfileParser()
	_, err := builder.Parse("org", "repo", "df_bad", "branch", nil)
	assert.NotNil(t, err, "Got non-nil output for mal-formed template action in df_bad")
}

func TestNestedVars(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "nested_var_df"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, _ := r.Parse("org", "repo", "nested_var_df", "branch", nil)

	const expected = `---
application: dinernotifications
globals:
  application: dinernotifications
pipelines:
- parameterConfig:
  - default: dinernotifications
	description: Service Name
	name: service
	required: true`

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
			expected: `
application: search
globals:
  type: foo
pipelines:
- foo: bar
  type: foo
- type: foobar`,
		},
		"df_global_nested": {
			filename: "df_global/nested",
			expected: `
application: search
globals:
  type: foo
pipelines:
- foo: bar
  type: foo
- type: foobar
`,
		},
		"df_global_appmodule": {
			filename: "df_app_global",
			expected: `
application: search
description: description
globals:
  type: foo
pipelines:
- foo: bar
  type: foo
- type: foobar
`,
		},
		"df_spec": {
			filename: "df_spec",
			expected: `
spec:
  name: search
  email: unknown@unknown.com
  dataSources:
    disabled: []
    enabled:
    - canaryConfigs
globals:
  type: foo
pipelines:
- foo: bar
  type: foo
- type: foobar`,
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			r := testDinghyfileParser()
			r.Builder.DinghyfileName = filepath.Base(c.filename)

			buf, _ := r.Parse("org", "repo", c.filename, "branch", nil)
			exp := strings.Join(strings.Fields(c.expected), "")
			actual := strings.Join(strings.Fields(buf.String()), "")
			assert.Equal(t, exp, actual)
		})
	}
}

func TestPreprocessingGlobalVars(t *testing.T) {
	input := `---
application: search
globals:
  system: order_tracking
pipelines:
- {{ module "preprod_deploy.pipeline.module" "application" "search" "master" "preprod" }}
- {{ module "prod_deploy.pipeline.module" "application" "search" "master" "prod" }}
`
	out, err := ParseGlobalVars(input)
	gvMap, ok := out.(map[string]interface{})
	assert.True(t, ok, "Something went wrong while extracting global vars")
	assert.Contains(t, gvMap, "system")
	assert.Equal(t, gvMap["system"], "order_tracking")
	assert.Nil(t, err)
}

//NOTE: this appears to be a noop in a yaml context. YAML doesn't necessarily need to "stringify" this way.
func TestPreprocess(t *testing.T) {
	cases := map[string]struct {
		input        string
		expected     string
		expectsError bool
	}{
		"main": {
			input:    `foo: {{ module "preprod_deploy.pipeline.module" "application" "test-spinnaker" "name" "deploy-preprod" "parameterConfig" [{"default": "selfserviceauthorization"}] }}`,
			expected: `foo: {{ module "preprod_deploy.pipeline.module" "application" "test-spinnaker" "name" "deploy-preprod" "parameterConfig" "[{\"default\": \"selfserviceauthorization\"}]" }}`,
		},
		"main2": {
			input:    `foo: {{ module "deployments_tracker_event.stage.module" "master" "preprod"  "parameters" {"build_user_email": "SPINNAKER", "publish_event_to_deployment_tracker": true, "region": "us_east", "service_name": "${parameters.service}", "service_version": "${parameters.version}", "failure_reason": "SMOKE_TESTS", "deployment_event_value": "SOFT_FAILURE" } "stageEnabled" {"expression": "${ #stage('us-east-1 Deployment Test')['status'] != \"SUCCEEDED\"}", "type": "expression"} "requisiteStageRefIds"  ["deploytestuseast1"] "refId" "dtraktestfailureuseast1" "name" "us-east-1 Deployment Tests Failure" }}`,
			expected: `foo: {{ module "deployments_tracker_event.stage.module" "master" "preprod" "parameters" "{\"build_user_email\": \"SPINNAKER\", \"publish_event_to_deployment_tracker\": true, \"region\": \"us_east\", \"service_name\": \"${parameters.service}\", \"service_version\": \"${parameters.version}\", \"failure_reason\": \"SMOKE_TESTS\", \"deployment_event_value\": \"SOFT_FAILURE\" }" "stageEnabled" "{\"expression\": \"${ #stage('us-east-1 Deployment Test')['status'] != \\\"SUCCEEDED\\\"}\", \"type\": \"expression\"}" "requisiteStageRefIds" "[\"deploytestuseast1\"]" "refId" "dtraktestfailureuseast1" "name" "us-east-1 Deployment Tests Failure" }}`,
		},
		"elvis": {
			input:    `foo: {{ var "asdf" ?: 42 }}`,
			expected: `foo: {{ var "asdf" 42 }}`,
		},
		"expectsError": {
			input:        `foo: {{ module "asdf" }`,
			expected:     "",
			expectsError: true,
		},
		"multiLineAction": {
			input: `
				application: maskednumbers
				pipelines:
                - {{
					module "integrate.pipeline.module"
						"application" "maskednumbers"
						"job" "platform/job/maskednumbers/job/maskednumbers_integrate"
						"triggerApp" "maskednumbers"
						"triggerPipeline" "deploy-preprod"
					}}`,
			expected: `
				application: maskednumbers
				pipelines:
                - {{ module "integrate.pipeline.module" "application" "maskednumbers" "job" "platform/job/maskednumbers/job/maskednumbers_integrate" "triggerApp" "maskednumbers" "triggerPipeline" "deploy-preprod" }}`,
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			a, err := preprocessor.Preprocess(c.input)

			if err != nil && c.expectsError {
				return
			} else if err != nil && !c.expectsError {
				t.Fatalf("got an error but didn't expect one: %s", err.Error())
			}
			assert.Equal(t, c.expected, a)
		})
	}
}

func TestSimpleWaitStage(t *testing.T) {
	r := testDinghyfileParser()
	buf, _ := r.Parse("org", "repo", "df3", "branch", nil)

	const expected = `
stages:
- name: Wait
  refId:
    c: d
  requisiteStageRefIds: ["1", "2", "3"]
  type: wait
  waitTime: 10
`

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

func TestSpillover(t *testing.T) {
	r := testDinghyfileParser()
	buf, _ := r.Parse("org", "repo", "df", "branch", nil)

	const expected = `
stages:
- foo: bar
  type: deploy
- type: jenkins
`

	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

type testStruct struct {
	Foo    string `yaml:"foo"`
	A      string `yaml:"a"`
	Nested struct {
		WaitTime int `yaml:"waitTime"`
	} `yaml:"nested"`
}

func TestModuleVariableSubstitution(t *testing.T) {
	r := testDinghyfileParser()
	ts := testStruct{}
	ret, err := r.Parse("org", "repo", "df2", "branch", nil)
	err = yaml.Unmarshal(ret.Bytes(), &ts)
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
	client.EXPECT().GetPipelines(gomock.Eq("triggerApp")).Return([]plank.Pipeline{plank.Pipeline{ID: "pipelineID", Name: "triggerPipeline"}}, nil).Times(1)
	r.Builder.Client = client

	vars := []dinghyfile.VarMap{
		{"triggerApp": "triggerApp", "triggerPipeline": "triggerPipeline"},
	}
	idFunc := r.pipelineIDFunc(vars).(func(string, string) (string, error))
	result, _ := idFunc("triggerApp", "triggerPipeline")
	assert.Equal(t, "pipelineID", result)
}

func TestPipelineIDFuncDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetPipelines(gomock.Eq("triggerApp")).Return(nil, errors.New("fake not found")).Times(1)
	r.Builder.Client = client

	vars := []dinghyfile.VarMap{
		{"triggerApp": "triggerApp"}, {"triggerPipeline": "triggerPipeline"},
	}
	idFunc := r.pipelineIDFunc(vars).(func(string, string) (string, error))
	result, _ := idFunc("triggerApp", "triggerPipeline")
	assert.Equal(t, "", result)
}

func TestPipelineIDRender(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := testDinghyfileParser()

	client := NewMockPlankClient(ctrl)
	client.EXPECT().GetPipelines(gomock.Eq("triggerApp")).Return([]plank.Pipeline{plank.Pipeline{ID: "pipelineID", Name: "triggerPipeline"}}, nil).Times(1)
	r.Builder.Client = client

	expected := `
application: pipelineidexample
failPipeline: true
name: Pipeline
pipeline: pipelineID
refId: 1
requisiteStageRefIds: []
type: pipeline
waitForCompletion: true`

	ret, err := r.Parse("org", "repo", "pipelineIDTest", "branch", nil)
	assert.Nil(t, err)
	assert.Equal(t, expected, ret.String())
}

func TestModuleEmptyString(t *testing.T) {
	r := testDinghyfileParser()
	ret, _ := r.Parse("org", "repo", "df4", "branch", nil)
	assert.Equal(t, `foo: ""`, ret.String())
}

func TestDeepVars(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "deep_var_df"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, _ := r.Parse("org", "repo", "deep_var_df", "branch", nil)

	const expected = `
application: dinernotifications
globals:
  application: dinernotifications
pipelines:
- parameterConfig:
  - description: Service Name
    name: service
    required: true
- parameterConfig:
  - artifact: artifact11
- parameterConfig:
  - artifact: artifact22
`
	// strip whitespace from both strings for assertion
	exp := strings.Join(strings.Fields(expected), "")
	actual := strings.Join(strings.Fields(buf.String()), "")
	assert.Equal(t, exp, actual)
}

func TestEmptyDefaultVar(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "deep_var_df"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, _ := r.Parse("org", "repo", "empty_default_variables", "branch", nil)

	const expected = `
application: dinernotifications
pipelines:

- parameterConfig:
  - default: ""
    description: Service Name
    name: service
    required: true

`

	exp := expected
	actual := buf.String()
	assert.Equal(t, exp, actual)
}

func TestConditionalArgs(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "if_params.dinghyfile"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"
	buf, err := r.Parse("org", "repo", "if_params.dinghyfile", "branch", nil)
	require.Nil(t, err)

	const raw = `
test: if_params
result:
- foo: bar
  biff: [ "foo", "bar" ]
`
	var expected interface{}
	err = yaml.Unmarshal([]byte(raw), &expected)
	require.Nil(t, err)
	expected_str, err := yaml.Marshal(expected)
	require.Nil(t, err)

	var actual interface{}
	err = yaml.Unmarshal(buf.Bytes(), &actual)
	require.Nil(t, err)
	actual_str, err := yaml.Marshal(actual)
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

	buf, err := r.Parse("org", "repo", "var_params.outer", "branch", nil)
	// Unfortunately, we don't currently catch this failure here.
	require.Nil(t, err)
	require.NotNil(t, buf)

	var actual interface{}
	err = yaml.Unmarshal(buf.Bytes(), &actual)
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

func TestPipelineSub(t *testing.T) {
	r := testDinghyfileParser()
	r.Builder.DinghyfileName = "pipeline_insert_base"
	r.Builder.TemplateOrg = "org"
	r.Builder.TemplateRepo = "repo"

	const expected = `
application: "foo"
pipelines:
- application: "foo"
  name: "bar"
  stages:
  - name: "stage"
    type: wait
  triggers: []
`

	buf, err := r.Parse("org", "repo", "pipeline_insert_base", "branch", nil)
	require.Nil(t, err)
	assert.Equal(t, expected, buf.String())
}
