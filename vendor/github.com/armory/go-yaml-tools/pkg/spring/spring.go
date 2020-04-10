package spring

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	yamlParse "gopkg.in/yaml.v2"

	"github.com/armory/go-yaml-tools/pkg/yaml"
)

var defaultConfigDirs = buildConfigDirs()

var defaultProfiles = []string{
	"armory",
	"local",
}

func buildConfigDirs() []string {
	paths := []string{
		// The order matters.
		"/home/spinnaker/config",
		"/opt/spinnaker/config",
		"/root/config",
	}
	usr, err := user.Current()
	if err == nil {
		paths = append(paths, filepath.Join(usr.HomeDir, ".spinnaker"))
	}
	return paths
}

// Use afero to create an abstraction layer between our package and the
// OS's file system. This will allow us to test our package.
var fs = afero.NewOsFs()

func loadConfig(configFile string) map[interface{}]interface{} {
	s := map[interface{}]interface{}{}
	if _, err := fs.Stat(configFile); err == nil {
		bytes, err := afero.ReadFile(fs, configFile)
		if err != nil {
			log.Errorf("Unable to open config file %s: %v", configFile, err)
			return nil
		}
		if err = yamlParse.Unmarshal(bytes, &s); err != nil {
			log.Errorf("Unable to parse config file %s: %v", configFile, err)
			return s
		}
		log.Info("Configured with settings from file: ", configFile)
	} else {
		log.Info("Config file ", configFile, " not present; falling back to default settings")
	}
	return s
}

//LoadProperties tries to do what spring properties manages by loading files
//using the right precendence and returning a merged map that contains all the
//keys and their values have been substituted for the correct value
//
//Usage:
//  If you want to load the following files:
//    - spinnaker.yml/yaml
//    - spinnaker-local.yml/yaml
//    - gate-local.yml/yaml
//    - gate-armory.yml/yaml
// Then For propNames you would give:
//	  ["spinnaker", "gate"]
// and you'll need to make sure your envKeyPairs has the following key pair as one of it's variables
//    SPRING_PROFILES_ACTIVE="armory,local"
func LoadProperties(propNames []string, configDir string, envKeyPairs []string) (map[string]interface{}, error) {
	envMap := keyPairToMap(envKeyPairs)
	profStr := envMap["SPRING_PROFILES_ACTIVE"]
	profs := strings.Split(profStr, ",")
	return loadProperties(propNames, configDir, profs, envMap)
}

// LoadDefault will use the following defaults:
//
// Check for the following config locations:
//  1. /opt/spinnaker/config
//  2. /home/spinnaker/config
//  3. /root/config
//  4. HOME_DIR/.spinnaker
// Profiles:
//  - armory
//  - local
// Env:
//  Pulls os.Environ()
//
// Specify propNames in the same way as LoadProperties().
func LoadDefault(propNames []string) (map[string]interface{}, error) {
	profs := profiles()
	dir := configDirectory()
	if dir == "" {
		return nil, errors.New("could not find config directory")
	}
	env := keyPairToMap(os.Environ())
	return loadProperties(propNames, dir, profs, env)
}

func keyPairToMap(keyPairs []string) map[string]string {
	m := map[string]string{}
	for _, keyPair := range keyPairs {
		split := strings.Split(keyPair, "=")
		m[split[0]] = split[1]
	}
	return m
}

func profiles() []string {
	s := os.Getenv("SPRING_PROFILES_ACTIVE")
	if len(s) > 0 {
		return strings.Split(s, ",")
	}
	return defaultProfiles
}

func configDirectory() string {
	for _, dir := range defaultConfigDirs {
		if _, err := fs.Stat(dir); err == nil {
			return dir
		}
	}
	return ""
}

func loadProperties(propNames []string, confDir string, profiles []string, envMap map[string]string) (map[string]interface{}, error) {
	propMaps := []map[interface{}]interface{}{}
	//first load the main props, i.e. gate.yml/yaml with no profile extensions
	for _, prop := range propNames {
		// yaml is "official"
		filePath := fmt.Sprintf("%s/%s.yaml", confDir, prop)
		config := loadConfig(filePath)

		// but people also use "yml" too, if we don't get anything let's try this
		if len(config) == 0 {
			filePath = fmt.Sprintf("%s/%s.yml", confDir, prop)
			config = loadConfig(filePath)
		}

		propMaps = append(propMaps, config)
	}

	for _, prop := range propNames {
		//we traverse the profiles array backwards for correct precedence
		//for i := len(profiles) - 1; i >= 0; i-- {
		for i := range profiles {
			p := profiles[i]
			pTrim := strings.TrimSpace(p)
			filePath := fmt.Sprintf("%s/%s-%s.yml", confDir, prop, pTrim)
			propMaps = append(propMaps, loadConfig(filePath))
		}
	}
	m, err := yaml.Resolve(propMaps, envMap)
	return m, err
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// Int64 is a helper routine that allocates a new int64 value
// to store v and returns a pointer to it.
func Int64(v int64) *int64 { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
