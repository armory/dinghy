package local

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
)

const (
	//LocalConfigSource is a variable of type string
	LocalConfigSource = "LocalSource"
)

//LocalSource is file source
type LocalSource struct {
	Configs *global.Settings
}

//NewLocalSource creates a source which can handler local settings
func NewLocalSource() *LocalSource {
	localConfigSource := new(LocalSource)
	return localConfigSource
}

//LoadSetupSettings load setttings for dinghy start-up
func (lSource *LocalSource) LoadSetupSettings() (*global.Settings, error) {

	initializer := source.NewInitialize()
	config, err := initializer.Autoconfigure()

	if err != nil {
		return nil, err
	}

	lSource.Configs = config

	return config, nil
}

func (*LocalSource) GetSourceName() string {
	return LocalConfigSource
}

//GetSettings get settings given the key
func (lSource *LocalSource) GetSettings(key string) (*global.Settings, error) {
	return lSource.Configs, nil
}
