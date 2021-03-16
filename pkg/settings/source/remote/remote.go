package remote

import (
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
)

const (
	//RemoteConfigSourceConst is a variable of type string
	RemoteConfigSourceConst = "RemoteSource"
)

//RemoteSource is file source
type RemoteSource struct {
}

//NewRemoteSource creates a source which can handler local files
func NewRemoteSource() *RemoteSource {
	remoteConfigSource := new(RemoteSource)
	return remoteConfigSource
}

func (rSource *RemoteSource) Load() (*global.Settings, error) {

	initializer := source.NewInitialize()
	config, err := initializer.Autoconfigure()

	if err != nil {
		return nil, err
	}

	return config, nil
}

//GetConfigurationByKey get one key value
func (rSource *RemoteSource) GetConfigurationByKey(key source.SettingField) interface{} {
	return nil
}

//GetStringByKey get one key value (string)
func (rSource *RemoteSource) GetStringByKey(key source.SettingField) string {
	return ""
}

//GetStringByKey get one key value (string)
func (rSource *RemoteSource) GetBoolByKey(key source.SettingField) bool {
	return false
}

//GetSourceName get name of source
func (*RemoteSource) GetSourceName() string {
	return RemoteConfigSourceConst
}

//GetStringArrayByKey get one key value (string)
func (rSource *RemoteSource) GetStringArrayByKey(key source.SettingField) []string {

	return nil
}