package dinghyfile

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"

	"text/template"

	"github.com/armory-io/dinghy/pkg/preprocessor"
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

// TODO: this function errors, it should be returning the error to the caller to be handled
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

			// checks for deepvariables, passes all the way down values from dinghyFile to module inside module
			deepVariable, ok := vars[i+1].(string)
			if ok && len(deepVariable) > 6 {
				if deepVariable[0:5] == "{{var" {
					for _, vm := range allVars {
						if val, exists := vm[deepVariable[6:len(deepVariable)-2]]; exists {
							log.Info("Substituting deepvariable ", vars[i], " : old value : ", deepVariable, " for new value: ", renderValue(val).(string))
							vars[i+1] = parseValue(val)
						}
					}
				}
			}
			newVars[key] = parseValue(vars[i+1])
		}

		result, _ := b.Render(b.TemplateOrg, b.TemplateRepo, mod, append([]varMap{newVars}, allVars...))
		return result.String()
	}
}

// TODO: this function errors, it should be returning the error to the caller to be handled
func pipelineIDFunc(b *PipelineBuilder, vars []varMap) interface{} {
	return func(app, pipelineName string, defaultVal ...interface{}) string {
		for _, vm := range vars {
			if val, exists := vm["triggerApp"]; exists {
				app = renderValue(val).(string)
				log.Info("Substituting pipeline trigger appname: ", app)
			}
			if val, exists := vm["triggerPipeline"]; exists {
				pipelineName = renderValue(val).(string)
				log.Info("Substituting pipeline trigger appname: ", app)
			}
		}
		pipeline, err := b.Client.GetPipeline(app, pipelineName)
		if err != nil {
			log.Errorf("could not get pipeline id for app %s, pipeline %s, err = %v", app, pipelineName, err)
		}
		return pipeline.ID
	}
}

// TODO: this function errors, it should be returning the error to the caller to be handled
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
			s, isStr := defaultVal[0].(string)

			if (isStr && len(s) > 0 ){
				if s[0] == '@' {
					// handle the case where the default value is another variable
					// see ENG-1921 for use case. e.g.,:
					// {{ var "servicename" ?: "@application" }}

					nested := s[1:]
					for _, vm := range vars {
						if val, exists := vm[nested]; exists {
							log.Info("Substituting nested variable: ", nested, ", val: ", val)
							return renderValue(val)
						}
					}
				}
			}
			return defaultVal[0]
		}
		return ""
	}
}

// Render renders the template
func (b *PipelineBuilder) Render(org, repo, path string, vars []varMap) (*bytes.Buffer, error) {
	deps := make(map[string]bool)

	// Download the template being rendered.
	contents, err := b.Downloader.Download(org, repo, path)
	if err != nil {
		return nil, err
	}

	// Preprocess to stringify any json args in calls to modules.
	contents, err = preprocessor.Preprocess(contents)
	if err != nil {
		return nil, err
	}

	// Extract global vars if we're processing a dinghyfile (and not a module)
	if filepath.Base(path) == b.DinghyfileName {
		gvs, err := preprocessor.ParseGlobalVars(contents)
		if err != nil {
			return nil, err
		}

		gvMap, ok := gvs.(map[string]interface{})
		if !ok {
			return nil, errors.New("Could not extract global vars")
		} else if len(gvMap) > 0 {
			vars = append(vars, gvMap)
		} else {
			log.Info("No global vars found in dinghyfile")
		}
	}

	funcMap := template.FuncMap{
		"module":     moduleFunc(b, org, deps, vars),
		"appModule":  moduleFunc(b, org, deps, vars),
		"pipelineID": pipelineIDFunc(b, vars),
		"var":        varFunc(vars),
	}

	// Parse the downloaded template.
	tmpl, err := template.New("dinghy-render").Funcs(funcMap).Parse(contents)
	if err != nil {
		return nil, err
	}

	// Run the template to verify the output.
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, "")
	if err != nil {
		return nil, err
	}

	// Record the dependencies we ran into.
	depUrls := make([]string, 0)
	for dep := range deps {
		depUrls = append(depUrls, dep)
	}
	b.Depman.SetDeps(b.Downloader.EncodeURL(org, repo, path), depUrls)

	return buf, nil
}
