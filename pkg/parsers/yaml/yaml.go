package yaml

import (
	"encoding/json"
	yaml2json "github.com/ghodss/yaml"
)

type DinghyYaml struct {}

func (d DinghyYaml) Unmarshal(data []byte, i interface{}) error {
	// convert our thing into something json compatible
	bytes, err := yaml2json.YAMLToJSON(data)
	if err != nil {
		return err
	}
	// convert json into...whatever interface
	if err := json.Unmarshal(bytes, i); err != nil {
		return err
	}
	return nil
}