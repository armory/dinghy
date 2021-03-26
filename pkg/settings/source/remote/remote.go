package remote

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
)

const (
	//RemoteConfigSource is a variable of type string
	RemoteConfigSource = "RemoteSource"
)

//RemoteSource is file source
type RemoteSource struct {
}

//NewRemoteSource creates a source which can handler remote settings
func NewRemoteSource() *RemoteSource {
	remoteConfigSource := new(RemoteSource)
	return remoteConfigSource
}

//LoadSetupSettings load setttings for dinghy start-up
func (lSource *RemoteSource) LoadSetupSettings() (*global.Settings, error) {

	initializer := source.NewInitialize()
	config, err := initializer.Autoconfigure()

	if err != nil {
		return nil, err
	}

	config.SQL.Enabled = true
	config.SQL.EventLogsOnly = false

	return config, nil
}

//GetSourceName get name of source
func (*RemoteSource) GetSourceName() string {
	return RemoteConfigSource
}

//GetSettings get settings given the key
func (lSource *RemoteSource) GetSettings(key string) (*global.Settings, error) {
	return nil, nil
}
