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

package preprocessor

import (
	"github.com/armory/dinghy/pkg/git"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreprocess(t *testing.T) {
	cases := map[string]struct {
		input        string
		expected     string
		expectsError bool
	}{
		"main": {
			input:    `{"foo": {{ module "preprod_deploy.pipeline.module" "application" "test-spinnaker" "name" "deploy-preprod" "parameterConfig" [{"default": "selfserviceauthorization"}] }}}`,
			expected: `{"foo": {{ module "preprod_deploy.pipeline.module" "application" "test-spinnaker" "name" "deploy-preprod" "parameterConfig" "[{\"default\": \"selfserviceauthorization\"}]" }}}`,
		},
		"main2": {
			input:    `{"foo": {{ module "deployments_tracker_event.stage.module" "master" "preprod"  "parameters" {"build_user_email": "SPINNAKER", "publish_event_to_deployment_tracker": true, "region": "us_east", "service_name": "${parameters.service}", "service_version": "${parameters.version}", "failure_reason": "SMOKE_TESTS", "deployment_event_value": "SOFT_FAILURE" } "stageEnabled" {"expression": "${ #stage('us-east-1 Deployment Test')['status'] != \"SUCCEEDED\"}", "type": "expression"} "requisiteStageRefIds"  ["deploytestuseast1"] "refId" "dtraktestfailureuseast1" "name" "us-east-1 Deployment Tests Failure" }}}`,
			expected: `{"foo": {{ module "deployments_tracker_event.stage.module" "master" "preprod" "parameters" "{\"build_user_email\": \"SPINNAKER\", \"publish_event_to_deployment_tracker\": true, \"region\": \"us_east\", \"service_name\": \"${parameters.service}\", \"service_version\": \"${parameters.version}\", \"failure_reason\": \"SMOKE_TESTS\", \"deployment_event_value\": \"SOFT_FAILURE\" }" "stageEnabled" "{\"expression\": \"${ #stage('us-east-1 Deployment Test')['status'] != \\\"SUCCEEDED\\\"}\", \"type\": \"expression\"}" "requisiteStageRefIds" "[\"deploytestuseast1\"]" "refId" "dtraktestfailureuseast1" "name" "us-east-1 Deployment Tests Failure" }}}`,
		},
		"elvis": {
			input:    `{"foo": {{ var "asdf" ?: 42 }}}`,
			expected: `{"foo": {{ var "asdf" 42 }}}`,
		},
		"expectsError": {
			input: `{
				"foo": {{ module "asdf" }
				}`,
			expected:     "",
			expectsError: true,
		},
		"multiLineAction": {
			input: `{
				"application": "maskednumbers",
				"pipelines": [
					{{
					module "integrate.pipeline.module"
						"application" "maskednumbers"
						"job" "platform/job/maskednumbers/job/maskednumbers_integrate"
						"triggerApp" "maskednumbers"
						"triggerPipeline" "deploy-preprod"
					}}
				]
				}`,
			expected: `{
				"application": "maskednumbers",
				"pipelines": [
					{{ module "integrate.pipeline.module" "application" "maskednumbers" "job" "platform/job/maskednumbers/job/maskednumbers_integrate" "triggerApp" "maskednumbers" "triggerPipeline" "deploy-preprod" }}
				]
				}`,
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			a, err := Preprocess(c.input)

			if err != nil && c.expectsError {
				return
			} else if err != nil && !c.expectsError {
				t.Fatalf("got an error but didn't expect one: %s", err.Error())
			}
			assert.Equal(t, c.expected, a)
		})
	}
}

func TestPreprocessingGlobalVars(t *testing.T) {
	input := `{
		"application": "search",
		"globals": {
			"system": "order_tracking"
		},
		"pipelines": [
			{{ module "preprod_deploy.pipeline.module" "application" "search" "master" "preprod" }},
			{{ module "prod_deploy.pipeline.module" "application" "search" "master" "prod" }}
	    ]
	  }`

	out, err := ParseGlobalVars(input, git.GitInfo{})
	gvMap, ok := out.(map[string]interface{})
	assert.True(t, ok, "Something went wrong while extracting global vars")
	assert.Contains(t, gvMap, "system")
	assert.Equal(t, gvMap["system"], "order_tracking")
	assert.Nil(t, err)
}

func TestContentShouldBeParsedCorrectly(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "content should be parsed correctly",
			args: args{
				content: `{"app": "myapp"}`,
			},
			wantErr: false,
		},
		{
			name: "parse should fail, content contains a invalid JSON",
			args: args{
				content: `{"myvalue": non-formatted  }`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ContentShouldBeParsedCorrectly(tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("ContentShouldBeParsedCorrectly() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dummyKV(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should return a dummy KV",
			want: `"a": "b"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dummyKV(tt.args.args...); got != tt.want {
				t.Errorf("dummyKV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dummyVar(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should return a dummy Var",
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dummyVar(tt.args.args...); got != tt.want {
				t.Errorf("dummyVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dummySlice(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should return a dummy Slice",
			want: make([]string, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dummySlice(tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dummySlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
