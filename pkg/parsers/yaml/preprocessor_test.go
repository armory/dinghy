package yaml

import (
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