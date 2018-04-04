package preprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreprocess(t *testing.T) {
	input := `{"foo": {{ module "preprod_deploy.pipeline.module" "application" "test-spinnaker" "name" "deploy-preprod" "parameterConfig" [{"default": "selfserviceauthorization"}] }}}`
	expected := `{"foo": {{ module "preprod_deploy.pipeline.module" "application" "test-spinnaker" "name" "deploy-preprod" "parameterConfig" "[{\"default\": \"selfserviceauthorization\"}]" }}}`

	assert.Equal(t, expected, Preprocess(input))
}

func TestPreprocessAgain(t *testing.T) {
	input := `{"foo": {{ module "deployments_tracker_event.stage.module" "master" "preprod"  "parameters" {"build_user_email": "SPINNAKER", "publish_event_to_deployment_tracker": true, "region": "us_east", "service_name": "${parameters.service}", "service_version": "${parameters.version}", "failure_reason": "SMOKE_TESTS", "deployment_event_value": "SOFT_FAILURE" } "stageEnabled" {"expression": "${ #stage('us-east-1 Deployment Test')['status'] != \"SUCCEEDED\"}", "type": "expression"} "requisiteStageRefIds"  ["deploytestuseast1"] "refId" "dtraktestfailureuseast1" "name" "us-east-1 Deployment Tests Failure" }}}`
	expected := `{"foo": {{ module "deployments_tracker_event.stage.module" "master" "preprod"  "parameters" "{\"build_user_email\": \"SPINNAKER\", \"publish_event_to_deployment_tracker\": true, \"region\": \"us_east\", \"service_name\": \"${parameters.service}\", \"service_version\": \"${parameters.version}\", \"failure_reason\": \"SMOKE_TESTS\", \"deployment_event_value\": \"SOFT_FAILURE\" }" "stageEnabled" "{\"expression\": \"${ #stage('us-east-1 Deployment Test')['status'] != \\\"SUCCEEDED\\\"}\", \"type\": \"expression\"}" "requisiteStageRefIds"  "[\"deploytestuseast1\"]" "refId" "dtraktestfailureuseast1" "name" "us-east-1 Deployment Tests Failure" }}}`

	assert.Equal(t, expected, Preprocess(input))
}

func TestElvisOperator(t *testing.T) {
	input := `{"foo": {{ var "asdf" ?: 42 }}}`
	expected := `{"foo": {{ var "asdf" 42 }}}`

	assert.Equal(t, expected, Preprocess(input))
}
