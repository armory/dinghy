package local

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/armory/dinghy/pkg/util"
	log "github.com/sirupsen/logrus"
	logr "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewLocalSource(t *testing.T) {
	i := NewLocalSource()

	var s source.SourceConfiguration = i
	assert.NotNil(t, i)
	assert.NotNil(t, s)
}

func TestLocalSource_Load(t *testing.T) {
	localSource := loadTestData()
	logrus := logr.New()
	config, err := localSource.LoadSetupSettings(logrus)

	assert.Nil(t, err)
	assert.Equal(t, LocalConfigSource, localSource.GetSourceName())
	assert.NotNil(t, config.SQL)
	assert.Equal(t, "sample-org", config.TemplateOrg)
	assert.Equal(t, false, config.SQL.Enabled)
}

func TestLocalSource_GetConfigurationByKey(t *testing.T) {
	localSource := loadTestData()
	logrus := logr.New()
	_, _ = localSource.LoadSetupSettings(logrus)
	r := new (http.Request)
	config, _, err := localSource.GetSettings(r, logrus)

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, global.SpinnakerService{
		Enabled: "true",
		BaseURL: "http://localhost:9000",
	}, config.Deck)
}

func loadTestData() *LocalSource {
	err := util.CopyToLocalSpinnaker("../testdata/dinghy-local.yml", "dinghy-local.yml")
	log.Error(err)
	return NewLocalSource()
}
