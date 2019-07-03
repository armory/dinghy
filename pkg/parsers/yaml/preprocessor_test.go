package yaml

import (
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

type TestYmlStruct struct {
	Foo string `yaml:"foo"`
	Bar []string `yaml:"bar"`
}

func TestUnmarshal(t *testing.T) {
	test := &TestYmlStruct{}
	data := `
foo: somedata
bar:
- baz
- fizz
`
	err := yaml.Unmarshal([]byte(data), test)
	assert.Nil(t, err)
	assert.Equal(t, "somedata", test.Foo)
}

func TestDinghyYmlUnmarshal(t *testing.T) {
	var d interface{}
	input := `
foo: somedata
bar:
- baz
- fizz
`
	err := Unmarshal([]byte(input), &d)
	assert.Nil(t, err)
	data, ok := d.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "somedata", data["foo"])
}

func TestDinghyYaml_Unmarshal_toDinghyfile(t *testing.T) {
	input := `
---
application: dinghyregression
pipelines:
- application: dinghyregression
  name: Dinghy Regression Yaml Test File
  appConfig: {}
  keepWaitingPipelines: false
  limitConcurrent: true
  stages:
  - name: Wait For It...
    refId: '1'
    requisiteStageRefIds: []
    type: wait
    waitTime: 4
  - name: Wait For It, Yo...
    refId: '2'
    requisiteStageRefIds: []
    type: wait
    waitTime: 20
  triggers: []
`

	um := &DinghyYaml{}
	danglyfile := dinghyfile.NewDinghyfile()
	err := um.Unmarshal([]byte(input), &danglyfile)
	assert.Nil(t, err)
	assert.Equal(t, "dinghyregression", danglyfile.Application)
	assert.Equal(t, 2, len(danglyfile.Pipelines[0].Stages))
}