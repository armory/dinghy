package preprocessor

import (
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

	out := ParseGlobalVars(input)
	gvMap, ok := out.(map[string]interface{})
	assert.True(t, ok, "Something went wrong while extracting global vars")
	assert.Contains(t, gvMap, "system")
	assert.Equal(t, gvMap["system"], "order_tracking")
}
