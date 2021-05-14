package settings

import (
	"flag"
	"github.com/armory/dinghy/pkg/settings/source/local"
	"github.com/armory/dinghy/pkg/settings/source/remote"
	logr "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadSettingsSingleTenant(t *testing.T) {
	flag.Set("mode", string(SingleTenant))
	log := logr.New()
	configuration, err := LoadSettings(log)

	assert.Nil(t, err)
	assert.Equal(t, configuration.GetSourceName(), local.LocalConfigSource)
}

func TestLoadSettingsMultiTenant(t *testing.T) {
	flag.Set("mode", string(MultiTenant))
	log := logr.New()
	configuration, err := LoadSettings(log)

	assert.Nil(t, err)
	assert.Equal(t, configuration.GetSourceName(), remote.RemoteConfigSource)
}
