package local

import (
	"fmt"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/oleiade/reflections"
	"log"
)

const (
	//LocalConfigSourceConst is a variable of type string
	LocalConfigSourceConst = "LocalSource"
)

//LocalSource is file source
type LocalSource struct {
	Configs global.Settings
}

//NewLocalSource creates a source which can handler local files
func NewLocalSource() *LocalSource {
	memoryConfigSource := new(LocalSource)
	return memoryConfigSource
}

//Load loads configuration
func (lSource *LocalSource) Load() (*global.Settings, error) {

	initializer := source.NewInitialize()
	config, err := initializer.Autoconfigure()

	if err != nil {
		return nil, err
	}

	lSource.Configs = *config

	return config, nil
}

//GetSourceName get name of source
func (*LocalSource) GetSourceName() string {
	return LocalConfigSourceConst
}

//GetConfigurationByKey get one key value
func (lSource *LocalSource) GetConfigurationByKey(key source.SettingField) interface{} {
	v, err := reflections.GetField(lSource.Configs, key.String())
	if err != nil {
		log.Fatal(err)
	}
	return v
}

//GetStringByKey get one key value (string)
func (lSource *LocalSource) GetStringByKey(key source.SettingField) string {
	v := lSource.GetConfigurationByKey(key)
	return fmt.Sprintf("%v", v)
}

//GetStringByKey get one key value (string)
func (lSource *LocalSource) GetBoolByKey(key source.SettingField) bool {
	v := lSource.GetConfigurationByKey(key)
	return v.(bool)
}

//GetStringArrayByKey get one key value (string)
func (lSource *LocalSource) GetStringArrayByKey(key source.SettingField) []string {
	v := lSource.GetConfigurationByKey(key)
	return v.([]string)
}
