package spring

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSubValues(t *testing.T) {
	wd, _ := os.Getwd()
	configDir := fmt.Sprintf("%s/../../test/springconf/", wd)

	propNames := []string{"spinnaker", "gate"}
	envPairs := []string{"SPRING_PROFILES_ACTIVE=armory,local"}
	props, _ := LoadProperties(propNames, configDir, envPairs)
	spinnaker := props["spinnaker"].(map[string]interface{})
	assert.Equal(t, "false", spinnaker["armory"])
	assert.Equal(t, "true", spinnaker["default"])
}

func TestDefaults(t *testing.T) {
	mockFs := afero.NewMemMapFs()
	mockFs.MkdirAll("/home/spinnaker/config", 0755)
	mockFile := `
spinnaker:
  something: false
`
	afero.WriteFile(mockFs, "/home/spinnaker/config/spinnaker-local.yml", []byte(mockFile), 0644)
	// Set the file system for the whole package. If we run into collisions we
	// might need to move this into a struct instead of keeping it at the
	// package level.
	fs = mockFs
	props, err := LoadDefault([]string{"spinnaker"})
	assert.Nil(t, err)
	spinnaker := props["spinnaker"].(map[string]interface{})
	assert.Equal(t, "false", spinnaker["something"])
}

func TestDefaultsWithMultipleProfiles(t *testing.T) {
	mockSpinnakerFile := `
services:
  front50:
    storage_bucket: mybucket2
`
	mockSpinnakerArmoryFile := `
services:
  front50:
    storage_bucket: mybucket
`
	mockFs := afero.NewMemMapFs()
	mockFs.MkdirAll("/home/spinnaker/config", 0755)
	afero.WriteFile(mockFs, "/home/spinnaker/config/spinnaker.yml", []byte(mockSpinnakerFile), 0644)
	afero.WriteFile(mockFs, "/home/spinnaker/config/spinnaker-armory.yml", []byte(mockSpinnakerArmoryFile), 0644)
	// Set the file system for the whole package. If we run into collisions we
	// might need to move this into a struct instead of keeping it at the
	// package level.
	fs = mockFs
	os.Setenv("SPINNAKER_DEFAULT_STORAGE_BUCKET", "mybucket2")
	os.Setenv("ARMORYSPINNAKER_CONF_STORE_BUCKET", "mybucket")
	props, err := LoadDefault([]string{"spinnaker"})
	t.Log(props)
	assert.Nil(t, err)
	type yaml struct {
		Services struct {
			Front50 struct {
				Bucket string `json:"storage_bucket" mapstructure:"storage_bucket"`
			} `json:"front50" mapstructure:"front50"`
		} `json:"services" mapstructure:"services"`
	}
	y := yaml{}
	mapstructure.WeakDecode(props, &y)
	assert.Equal(t, "mybucket", y.Services.Front50.Bucket)
}

func TestConfigDirs(t *testing.T) {
	assert.Equal(t, len(defaultConfigDirs), 4)
}
