package local

import (
	"fmt"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"github.com/oleiade/reflections"
	"log"
	"strconv"
)

const (
	//LocalConfigSource is a variable of type string
	LocalConfigSource = "LocalSource"
)

//LocalSource is file source
type LocalSource struct {
	Configs global.Settings
}

//NewLocalSource creates a source which can handler local files
func NewLocalSource() *LocalSource {
	localConfigSource := new(LocalSource)
	return localConfigSource
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
	return LocalConfigSource
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

//GetBoolByKey get one key value (bool)
func (lSource *LocalSource) GetBoolByKey(key source.SettingField) bool {
	v := lSource.GetStringByKey(key)
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Fatal(err)
	}
	return b

}

//GetStringArrayByKey get one key value ([]string)
func (lSource *LocalSource) GetStringArrayByKey(key source.SettingField) []string {
	v := lSource.GetConfigurationByKey(key)
	return v.([]string)
}
