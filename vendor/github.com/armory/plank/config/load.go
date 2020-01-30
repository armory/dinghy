package config

import (
	"github.com/armory/go-yaml-tools/pkg/spring"
	"github.com/mitchellh/mapstructure"
)

// Load settings from yaml file.
func Load(dest interface{}, services ...string) error {
	m, err := spring.LoadDefault(services)
	if err != nil {
		return err
	}
	err = mapstructure.WeakDecode(m, &dest)
	if err != nil {
		return err
	}
	return nil
}
