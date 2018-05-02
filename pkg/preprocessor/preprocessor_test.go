package preprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreprocess(t *testing.T) {
	input := `{"foo": {{ module "preprod_deploy.pipeline.module" "application" "test-spinnaker" "name" "deploy-preprod" "parameterConfig" [{"default": "selfserviceauthorization"}] }}}`
	expected := `{"foo": {{ module "preprod_deploy.pipeline.module" "application" "test-spinnaker" "name" "deploy-preprod" "parameterConfig" "[{\"default\": \"selfserviceauthorization\"}]" }}}`
	actual, _ := Preprocess(input)
	assert.Equal(t, expected, actual)
}

func TestPreprocessAgain(t *testing.T) {
	input := `{"foo": {{ module "deployments_tracker_event.stage.module" "master" "preprod"  "parameters" {"build_user_email": "SPINNAKER", "publish_event_to_deployment_tracker": true, "region": "us_east", "service_name": "${parameters.service}", "service_version": "${parameters.version}", "failure_reason": "SMOKE_TESTS", "deployment_event_value": "SOFT_FAILURE" } "stageEnabled" {"expression": "${ #stage('us-east-1 Deployment Test')['status'] != \"SUCCEEDED\"}", "type": "expression"} "requisiteStageRefIds"  ["deploytestuseast1"] "refId" "dtraktestfailureuseast1" "name" "us-east-1 Deployment Tests Failure" }}}`
	expected := `{"foo": {{ module "deployments_tracker_event.stage.module" "master" "preprod"  "parameters" "{\"build_user_email\": \"SPINNAKER\", \"publish_event_to_deployment_tracker\": true, \"region\": \"us_east\", \"service_name\": \"${parameters.service}\", \"service_version\": \"${parameters.version}\", \"failure_reason\": \"SMOKE_TESTS\", \"deployment_event_value\": \"SOFT_FAILURE\" }" "stageEnabled" "{\"expression\": \"${ #stage('us-east-1 Deployment Test')['status'] != \\\"SUCCEEDED\\\"}\", \"type\": \"expression\"}" "requisiteStageRefIds"  "[\"deploytestuseast1\"]" "refId" "dtraktestfailureuseast1" "name" "us-east-1 Deployment Tests Failure" }}}`
	actual, _ := Preprocess(input)
	assert.Equal(t, expected, actual)
}

func TestElvisOperator(t *testing.T) {
	input := `{"foo": {{ var "asdf" ?: 42 }}}`
	expected := `{"foo": {{ var "asdf" 42 }}}`
	actual, _ := Preprocess(input)
	assert.Equal(t, expected, actual)
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

func TestMalformedAction(t *testing.T) {
	input := `{
				"foo": {{ module "asdf" }
			  }`
	_, err := Preprocess(input)
	assert.Error(t, err, "Preprocessor didn't return error for malformed template action")
}
