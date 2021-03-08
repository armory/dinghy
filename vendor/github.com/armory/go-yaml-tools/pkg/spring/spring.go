package spring

import (
	"context"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	yamlParse "gopkg.in/yaml.v2"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/armory/go-yaml-tools/pkg/yaml"
)

type springEnv struct {
	defaultConfigDirs []string
	defaultProfiles   []string
	configDir         string
	envMap            map[string]string
}

func (s *springEnv) initialize() {
	s.buildConfigDirs()
	s.defaultProfiles = []string{"armory", "local"}
	s.configDir = s.configDirectory()
	s.envMap = keyPairToMap(os.Environ())
}

func (s *springEnv) buildConfigDirs() {
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
	s.defaultConfigDirs = paths
}

func (s *springEnv) configDirectory() string {
	for _, dir := range s.defaultConfigDirs {
		if _, err := fs.Stat(dir); err == nil {
			return dir
		}
	}
	return ""
}

func (s *springEnv) profiles() []string {
	p := os.Getenv("SPRING_PROFILES_ACTIVE")
	if len(p) > 0 {
		return strings.Split(p, ",")
	}
	return s.defaultProfiles
}

// Use afero to create an abstraction layer between our package and the
// OS's file system. This will allow us to test our package.
var fs = afero.NewOsFs()

func loadConfig(configFile string) (map[interface{}]interface{}, error) {
	s := map[interface{}]interface{}{}
	if _, err := fs.Stat(configFile); err == nil {
		bytes, err := afero.ReadFile(fs, configFile)
		if err != nil {
			log.Errorf("Unable to open config file %s: %v", configFile, err)
			return nil, nil
		}
		if err = yamlParse.Unmarshal(bytes, &s); err != nil {
			return s, fmt.Errorf("unable to parse config file %s: %w", configFile, err)
		}
		log.Info("Configured with settings from file: ", configFile)
	} else {
		logFsStatError(err, "Config file ", configFile, " not present; falling back to default settings")
	}
	return s, nil
}

func logFsStatError(err error, args ...interface{}) {
	if os.IsNotExist(err) {
		log.WithError(err).Debug(args...)
		return
	}
	log.WithError(err).Error(args...)
}

//LoadProperties tries to do what spring properties manages by loading files
//using the right precedence and returning a merged map that contains all the
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
	config, _, err := loadProperties(propNames, configDir, profs, envMap)
	return config, err
}

// Similar to LoadDefault but provides a callback function that will be invoked when a configuration change
// is detected. Parsing errors are also provided to the callback, so check for these as well.
// This works by keeping track of files parsed during the initial parsing, it means that files will only
// be tracked if they contain something. e.g. you cannot dynamically add a profile.
// Environment variables are frozen on the initial run. This is by design
func LoadDefaultDynamic(ctx context.Context, propNames []string, updateFn func(map[string]interface{}, error)) (map[string]interface{}, error) {
	env := springEnv{}
	env.initialize()
	return loadDefaultDynamicWithEnv(env, ctx, propNames, updateFn)
}

func loadDefaultDynamicWithEnv(env springEnv, ctx context.Context, propNames []string, updateFn func(map[string]interface{}, error)) (map[string]interface{}, error) {
	if env.configDir == "" {
		return nil, errors.New("could not find config directory")
	}

	config, files, err := loadProperties(propNames, env.configDir, env.profiles(), env.envMap)
	if len(files) > 0 {
		go watchConfigFiles(ctx, files, env.envMap, updateFn)
	}
	return config, err
}

func watchConfigFiles(ctx context.Context, files []string, envMap map[string]string, updateFn func(map[string]interface{}, error)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithError(err).Errorf("unable to watch any file")
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Debugf("file %s modified, rebuilding config", event.Name)
					var cfgs []map[interface{}]interface{}
					for _, f := range files {
						config, err := loadConfig(f)
						if err != nil {
							log.Errorf("file %s had error %s", f, err.Error())
						}
						cfgs = append(cfgs, config)
					}
					m, err := yaml.Resolve(cfgs, envMap)
					updateFn(m, err)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	for _, f := range files {
		if err = watcher.Add(f); err != nil {
			log.WithError(err).Errorf("unable to watch file changes for %s", f)
		}
	}
	<-ctx.Done()
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
	env := springEnv{}
	env.initialize()
	return loadDefaultWithEnv(env, propNames)
}

func loadDefaultWithEnv(env springEnv, propNames []string) (map[string]interface{}, error) {
	if env.configDir == "" {
		return nil, errors.New("could not find config directory")
	}
	config, _, err := loadProperties(propNames, env.configDir, env.profiles(), env.envMap)
	return config, err
}

func keyPairToMap(keyPairs []string) map[string]string {
	m := map[string]string{}
	for _, keyPair := range keyPairs {
		split := strings.Split(keyPair, "=")
		m[split[0]] = split[1]
	}
	return m
}

func loadProperties(propNames []string, confDir string, profiles []string, envMap map[string]string) (map[string]interface{}, []string, error) {
	var propMaps []map[interface{}]interface{}
	var filePaths []string
	//first load the main props, i.e. gate.yml/yaml with no profile extensions
	for _, prop := range propNames {
		// yaml is "official"
		config, filePath, err := loadPropertyFromFile(fmt.Sprintf("%s/%s", confDir, prop))
		// file might have been unparsable
		if err != nil {
			return nil, filePaths, err
		}
		if len(config) > 0 {
			propMaps = append(propMaps, config)
			filePaths = append(filePaths, filePath)
		}
	}

	for _, prop := range propNames {
		//we traverse the profiles array backwards for correct precedence
		//for i := len(profiles) - 1; i >= 0; i-- {
		for i := range profiles {
			p := profiles[i]
			pTrim := strings.TrimSpace(p)
			config, filePath, err := loadPropertyFromFile(fmt.Sprintf("%s/%s-%s", confDir, prop, pTrim))
			if err != nil {
				return nil, filePaths, err
			}
			if len(config) > 0 {
				propMaps = append(propMaps, config)
				filePaths = append(filePaths, filePath)
			}
		}
	}
	m, err := yaml.Resolve(propMaps, envMap)
	return m, filePaths, err
}

func loadPropertyFromFile(pathPrefix string) (map[interface{}]interface{}, string, error) {
	filePath := fmt.Sprintf("%s.yaml", pathPrefix)
	config, err := loadConfig(filePath)
	if err != nil {
		return config, filePath, err
	}
	if len(config) > 0 {
		return config, filePath, nil
	}

	// but people also use "yml" too, if we don't get anything let's try this
	filePath = fmt.Sprintf("%s.yml", pathPrefix)
	config, err = loadConfig(filePath)
	return config, filePath, err
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
