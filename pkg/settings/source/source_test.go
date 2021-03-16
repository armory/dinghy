package source

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Ensure_Test_Source_Implements_Source_Interface(t *testing.T) {
	var test Source = NewTestSource()
	assert.NotNil(t, test)
}

type TestSource struct {
}

func NewTestSource() *TestSource {
	testSource := new(TestSource)
	return testSource
}

func (*TestSource) Load() (*global.Settings, error) {
	return nil, nil
}

func (*TestSource) GetConfigurationByKey(key SettingField) interface{} {
	return nil
}

func (*TestSource) GetStringByKey(key SettingField) string {
	return ""
}

func (*TestSource) GetBoolByKey(key SettingField) bool {
	return false
}

func (*TestSource) GetSourceName() string {
	return ""
}

func (*TestSource) GetStringArrayByKey(key SettingField) []string {
	return nil
}
