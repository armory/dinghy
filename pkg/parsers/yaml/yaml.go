package yaml

import "gopkg.in/yaml.v2"

type DinghyYaml struct {}

func (d DinghyYaml) Unmarshal(data []byte, i interface{}) error {
	err := yaml.Unmarshal(data, i)
	if err != nil {
		return err
	}
	return nil
}