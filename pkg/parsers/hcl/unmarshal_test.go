package hcl

import (
	"encoding/json"
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/hashicorp/hcl"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestHclStruct struct {
	Foo string `hcl:"foo"`
	Bar []string `hcl:"bar"`
}

func TestUnmarshal(t *testing.T) {
	test := &TestHclStruct{}

	data := `
	foo = "bar"
	bar = [
		"data",
		"moredata",
	]
	`
	err := hcl.Unmarshal([]byte(data), test)
	assert.Nil(t, err)
	assert.Equal(t, "moredata", test.Bar[1])
}

// Make sure we can unmarshal HCL cleanly into Dinghyfile
func TestDinghyHcl_Unmarshal(t *testing.T) {
	inputType1 := `
			"application" = "search"

			"globals" = {
			  "system" = "order_tracking"
			}

			"pipelines" = {
				"type" = "test"
			}

			"pipelines" = {
				"type" = "baz"
			}
	`

	inputType2 := `
			"application" = "search"

			"globals" = {
			  "system" = "order_tracking"
			}

			"pipelines" = [
				{ "type" = "test" },
				{ "type" = "baz" }
			]
	`

	um := &DinghyHcl{}
	danglyfile := dinghyfile.NewDinghyfile()
	err := um.Unmarshal([]byte(inputType1), &danglyfile)
	assert.Nil(t, err)
	assert.Equal(t, "search", danglyfile.Application)
	assert.Equal(t, 2, len(danglyfile.Pipelines))

	danglyfile = dinghyfile.NewDinghyfile()
	err = um.Unmarshal([]byte(inputType2), &danglyfile)
	assert.Nil(t, err)
	assert.Equal(t, "search", danglyfile.Application)
	assert.Equal(t, 2, len(danglyfile.Pipelines))
}

func TestDinghyfileRegression(t *testing.T) {
	input := `
		"application" = "dinghyregression"
		
		"pipelines" = [
			{
			  "appConfig" = {}
			  "application" = "dinghyregression"
			  "keepWaitingPipelines" = false
			  "limitConcurrent" = true
			  "name" = "Dinghy Regression Test File (in HCL!)"
			  "stages" = [
				{
					"name" = "Wait For It...In HCL!!"
					"refId" = "1"
					"requisiteStageRefIds" = []
					"type" = "wait"
					"waitTime" = 5
				}
			  ]
			  "triggers" = {}
			}
		]
	`
	// convert the above hcl
	um := &DinghyHcl{}
	danglyfile := dinghyfile.NewDinghyfile()
	err := um.Unmarshal([]byte(input), &danglyfile)
	assert.Nil(t, err)
	assert.Equal(t, "dinghyregression", danglyfile.Application)

	// convert our HCL->json text to ensure we've parsed the objects correctly
	hclToJson, _ := json.Marshal(danglyfile)
	anotherDinghyfile := dinghyfile.NewDinghyfile()
	err = json.Unmarshal(hclToJson, &anotherDinghyfile)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(anotherDinghyfile.Pipelines))
	assert.Equal(t, 1, len(anotherDinghyfile.Pipelines[0].Stages))
}