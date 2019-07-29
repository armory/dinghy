package yaml

import (
	"encoding/json"
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

func TestDinghyYaml_Unmarshal_RemarshalToJson(t *testing.T) {
	input := `
application: "apptest"
# You can do inline comments now
globals:
  waitTime: 42
  retries: 5
pipelines: 
- application: "apptest"
  name: "Staging and Smoke Test  YAML"
  parameterConfig:
  - label: "Sha Shiv"
    name: "venv_sha"
    required: true
  - label: "Sha Tarball of GitHub Repo"
    name: "python_sha"
    required: true
  - label: "Branch to provision"
    name: "provision_branch"
    description: "Branch to provision"
    default: "master"
  stages:
  - name: Check Preconditions
    refId: '22'
    requisiteStageRefIds: []
    type: checkPreconditions
  - name: Find Base Image from Tags
    cloudProvider: aws
    cloudProviderType: aws
    packageName: something_python_bionic_
    refId: '20'
    tags:
      Name: something_python
      distro: ubuntu
      env: prod
      release: bionic
    regions:
      - us-west-1
    requisiteStageRefIds: 
      - '22'
    type: findImageFromTags
  - name: Bake AMI
    baseAmi: "${ #stage('Find Base Image from Tags')['context']['amiDetails'][0]['imageId']}"
    baseLabel: release
    baseOs: ubuntu
    cloudProviderType: aws
    package: none_${parameters["python_sha"]}
    refId: '21'
    region: us-west-1
    regions:
      - us-west-1
    requisiteStageRefIds:
      - '20'
    storeType: ebs
    templateFileName: tarball-shiv-install.json
    vmType: hvm
    type: bake
  triggers: []
`
	um := &DinghyYaml{}
	danglyfile := dinghyfile.NewDinghyfile()
	err := um.Unmarshal([]byte(input), &danglyfile)
	assert.Nil(t, err)
	stageName := danglyfile.Pipelines[0].Stages[0]["name"]
	assert.Equal(t, "Check Preconditions", stageName)
	_, err = json.Marshal(danglyfile)
	assert.Nil(t, err)
}