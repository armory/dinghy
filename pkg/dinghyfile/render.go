package dinghyfile

import (
	"bytes"
	"encoding/json"

	"text/template"

	"github.com/armory-io/dinghy/pkg/preprocessor"
	"github.com/armory-io/dinghy/pkg/spinnaker"
	log "github.com/sirupsen/logrus"
)

func parseValue(val interface{}) interface{} {
	if jsonStr, ok := val.(string); ok && len(jsonStr) > 0 {
		if jsonStr[0] == '{' {
			json.Unmarshal([]byte(jsonStr), &val)
		}
		if jsonStr[0] == '[' {
			json.Unmarshal([]byte(jsonStr), &val)
		}
	}

	return val
}

type varMap map[string]interface{}

func moduleFunc(b *PipelineBuilder, org string, deps map[string]bool, allVars []varMap) interface{} {
	return func(mod string, vars ...interface{}) string {
		// Record the dependency.
		child := b.Downloader.EncodeURL(org, b.TemplateRepo, mod)
		if _, exists := deps[child]; !exists {
			deps[child] = true
		}

		length := len(vars)
		if length%2 != 0 {
			log.Warnf("odd number of parameters received to module %s", mod)
		}

		newVars := make(varMap)
		for i := 0; i+1 < length; i += 2 {
			key, ok := vars[i].(string)
			if !ok {
				log.Errorf("dict keys must be strings in module: %s", mod)
				return ""
			}

			newVars[key] = parseValue(vars[i+1])
		}

		result := b.Render(b.TemplateOrg, b.TemplateRepo, mod, append([]varMap{newVars}, allVars...))
		return result.String()
	}
}

func pipelineIDFunc(app, pipelineName string) string {
	id, err := spinnaker.GetPipelineID(app, pipelineName)
	if err != nil {
		log.Errorf("could not get pipeline id for app %s, pipeline %s, err = %v", app, pipelineName, err)
	}
	return id
}

func renderValue(val interface{}) interface{} {
	// If it's an unserialized JSON array, serialize it back to JSON.
	if newval, ok := val.([]interface{}); ok {
		buf, err := json.Marshal(newval)
		if err != nil {
			log.Errorf("unable to json.marshal value %v", val)
			return ""
		}
		return string(buf)
	}

	// If it's an unserialized JSON object, serialize it back to JSON.
	if newval, ok := val.(map[string]interface{}); ok {
		buf, err := json.Marshal(newval)
		if err != nil {
			log.Errorf("unable to json.marshal value %v", val)
			return ""
		}
		return string(buf)
	}

	// Return value as is.
	return val
}

func varFunc(vars []varMap) interface{} {
	return func(varName string, defaultVal ...interface{}) interface{} {
		for _, vm := range vars {
			if val, exists := vm[varName]; exists {
				return renderValue(val)
			}
		}

		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return ""
	}
}

// Render renders the template
func (b *PipelineBuilder) Render(org, repo, path string, vars []varMap) *bytes.Buffer {
	deps := make(map[string]bool)

	funcMap := template.FuncMap{
		"module":     moduleFunc(b, org, deps, vars),
		"pipelineID": pipelineIDFunc,
		"var":        varFunc(vars),
	}

	// Download the template being rendered.
	contents, err := b.Downloader.Download(org, repo, path)
	if err != nil {
		log.Errorf("could not download %s/%s/%s", org, repo, path)
		return nil
	}

	// Preprocess to stringify any json args in calls to modules.
	contents = preprocessor.Preprocess(contents)

	// Parse the downloaded template.
	tmpl, err := template.New("dinghy-render").Funcs(funcMap).Parse(contents)
	if err != nil {
		log.Errorf("template parsing: %s", err)
		return nil
	}

	// Run the template to verify the output.
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, "")
	if err != nil {
		log.Errorf("template execution: %s", err)
		return nil
	}

	// Record the dependencies we ran into.
	depUrls := make([]string, 0)
	for dep := range deps {
		depUrls = append(depUrls, dep)
	}
	b.Depman.SetDeps(b.Downloader.EncodeURL(org, repo, path), depUrls)

	return buf
}
