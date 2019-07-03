package hcl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// HCL Arrays of Objects must have {} around them
func TestPreprocessingGlobalVars(t *testing.T) {

	input := `
			"application" = "search"

			"globals" = {
			  "system" = "order_tracking"
			}

			"pipelines" = [
				{ {{ module "preprod_deploy.pipeline.module" "application" "search" "master" "preprod" }} },
				{ {{ module "prod_deploy.pipeline.module" "application" "search" "master" "prod" }} }
			]
	`

	out, err := ParseGlobalVars(input)
	gvMap, ok := out.(map[string]interface{})
	assert.True(t, ok, "Something went wrong while extracting global vars")
	assert.Contains(t, gvMap, "system")
	assert.Equal(t, gvMap["system"], "order_tracking")
	assert.Nil(t, err)
}

// Second valid form for object arrays is to simply specify the object array key multiple times
func TestPreprocessingGlobalVarsObjectArray(t *testing.T) {

	input := `
			"application" = "search"

			"globals" = {
			  "system" = "order_tracking"
			}

			"pipelines" = {
				{{ module "preprod_deploy.pipeline.module" "application" "search" "master" "preprod" }}
			}

			"pipelines" = {
				{{ module "prod_deploy.pipeline.module" "application" "search" "master" "prod" }}
			}
	`

	out, err := ParseGlobalVars(input)
	gvMap, ok := out.(map[string]interface{})
	assert.True(t, ok, "Something went wrong while extracting global vars")
	assert.Contains(t, gvMap, "system")
	assert.Equal(t, gvMap["system"], "order_tracking")
	assert.Nil(t, err)
}
