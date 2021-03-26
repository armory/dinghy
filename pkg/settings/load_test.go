package settings

import (
	"flag"
	"github.com/armory/dinghy/pkg/settings/source/local"
	"github.com/armory/dinghy/pkg/settings/source/remote"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadSettingsSingleTenant(t *testing.T) {
	flag.Set("mode", string(SingleTenant))
	configuration, err := LoadSettings()

	assert.Nil(t, err)
	assert.Equal(t, configuration.GetSourceName(), local.LocalConfigSource)
}

func TestLoadSettingsMultiTenant(t *testing.T) {
	flag.Set("mode", string(MultiTenant))
	configuration, err := LoadSettings()

	assert.Nil(t, err)
	assert.Equal(t, configuration.GetSourceName(), remote.RemoteConfigSource)
}
