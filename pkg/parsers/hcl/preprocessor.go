package hcl

import (
	"bytes"
	"github.com/hashicorp/hcl"
	"text/template"
)

func dummySubstitute(args ...interface{}) string {
	return `"a" = "b"`
}

func dummyKV(args ...interface{}) string {
	return `"a" = "b"`
}

// since {{ var ... }} can be a string or an int!
func dummyVar(args ...interface{}) string {
	return "1"
}

func dummySlice(args ...interface{}) []string {
	return make([]string, 0)
}

// removeModules replaces all template function calls ({{ ... }}) in the dinghyfile with
// the YAML: a: b so that we can extract the global vars using Yaml.Unmarshal
func removeModules(input string) string {

	funcMap := template.FuncMap{
		"module":     dummySubstitute,
		"appModule":  dummyKV,
		"var":        dummyVar,
		"pipelineID": dummyVar,
		"makeSlice":  dummySlice,
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

func FlattenMap(data *map[string]interface{}) {
	for k, v := range *data {
		if mapArray, ok := v.([]map[string]interface{}); ok {
			if len(mapArray) == 1 {
				(*data)[k] = mapArray[0]
			}
		}
	}

	for _, v := range *data {
		if m, ok := v.(map[string]interface{}); ok {
			FlattenMap(&m)
		}
	}
}

// ParseGlobalVars returns the map of global variables in the dinghyfile
func ParseGlobalVars(input string) (interface{}, error) {

	d := make(map[string]interface{})
	input = removeModules(input)
	err := hcl.Unmarshal([]byte(input), &d)
	if err != nil {
		return nil, err
	}

	// HCL unmarsal/decode returns []map[string]interface{}. Flatten it.
	FlattenMap(&d)
	if val, ok := d["globals"]; ok {
		return val, nil
	}
	return make(map[string]interface{}), nil
}
