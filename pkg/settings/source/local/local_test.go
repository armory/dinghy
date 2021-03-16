package local

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/armory/dinghy/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLocalSource(t *testing.T) {
	i := NewLocalSource()

	var s source.Source = i
	assert.NotNil(t, i)
	assert.NotNil(t, s)
}

func TestLocalSource_Load(t *testing.T) {
	localSource := loadTestDate()

	config, err := localSource.Load()

	assert.Nil(t, err)
	assert.Equal(t, LocalConfigSource, localSource.GetSourceName())
	assert.NotNil(t, config.SQL)
	assert.Equal(t, "sample-org", config.TemplateOrg)
	assert.Equal(t, false, config.SQL.Enabled)
}

func TestLocalSource_GetConfigurationByKey(t *testing.T) {
	localSource := loadTestDate()
	_, _ = localSource.Load()
	config := localSource.GetConfigurationByKey(source.Deck).(global.SpinnakerService)

	assert.NotNil(t, config)
	assert.Equal(t, global.SpinnakerService{
		Enabled: "true",
		BaseURL: "http://localhost:9000",
	}, config)
}

func TestLocalSource_GetStringByKey(t *testing.T) {
	localSource := loadTestDate()
	_, _ = localSource.Load()
	gitHubToken := localSource.GetStringByKey(source.GitHubToken)

	assert.Equal(t, "xxxxxxxxxx", gitHubToken)
}

func TestLocalSource_GetBoolByKey(t *testing.T) {
	localSource := loadTestDate()
	_, _ = localSource.Load()
	autoLockPipelines := localSource.GetBoolByKey(source.AutoLockPipelines)

	assert.Equal(t, true, autoLockPipelines)
}

func TestLocalSource_GetStringArrayByKey(t *testing.T) {
	localSource := loadTestDate()
	_, _ = localSource.Load()
	webhookValidationEnabledProviders := localSource.GetStringArrayByKey(source.WebhookValidationEnabledProviders)

	assert.Equal(t, []string{"github"}, webhookValidationEnabledProviders)
}

func loadTestDate() *LocalSource {
	util.CopyToLocalSpinnaker("../testdata/dinghy-local.yml", "dinghy-local.yml")
	return NewLocalSource()
}
