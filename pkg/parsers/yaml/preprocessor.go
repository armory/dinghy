package yaml

import (
	"bytes"
	"fmt"
	y "gopkg.in/yaml.v2"
	"text/template"
)

// **** Shamelessly taken from https://github.com/go-yaml/yaml/issues/139#issuecomment-220072190
func Unmarshal(in []byte, out interface{}) error {
	var res interface{}

	if err := y.Unmarshal(in, &res); err != nil {
		return err
	}
	*out.(*interface{}) = cleanupMapValue(res)

	return nil
}

// Marshal YAML wrapper function.
func Marshal(in interface{}) ([]byte, error) {
	return y.Marshal(in)
}

func cleanupInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = cleanupMapValue(v)
	}
	return res
}

func cleanupInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// **** End shameless yaml to map[string]interface{} conversion

func dummySubstitute(args ...interface{}) string {
	return `a: b`
}

func dummyKV(args ...interface{}) string {
	return `a: b`
}

// since {{ var ... }} can be a string or an int!
func dummyVar(args ...interface{}) string {
	return "1"
}

// removeModules replaces all template function calls ({{ ... }}) in the dinghyfile with
// the YAML: a: b so that we can extract the global vars using Yaml.Unmarshal
func removeModules(input string) string {

	funcMap := template.FuncMap{
		"module":     dummySubstitute,
		"appModule":  dummyKV,
		"var":        dummyVar,
		"pipelineID": dummyVar,
	}

	tmpl, err := template.New("blank-out").Funcs(funcMap).Parse(input)
	if err != nil {
		return input
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, "")
	if err != nil {
		return input
	}

	return buf.String()
}


// ParseGlobalVars returns the map of global variables in the dinghyfile
func ParseGlobalVars(input string) (interface{}, error) {

	var d interface{}
	input = removeModules(input)
	err := Unmarshal([]byte(input), &d)
	if err != nil {
		return nil, err
	}
	data, ok := d.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to serialize yaml to map")
	}

	if val, ok := data["globals"]; ok {
		return val, nil
	}
	return make(map[string]interface{}), nil
}