package dinghyfile

import (
	"bytes"
	"encoding/json"
	"path/filepath"

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

			// checks for deepvariables, passes all the way down values from dinghyFile to module inside module
			if jsonStr, ok := vars[i+1].(string); ok && len(jsonStr) > 6 {
				if vars[i+1].(string)[0:5] == "{{var" {
					for _, vm := range allVars {
						if val, exists := vm[ vars[i+1].(string)[6:len(vars[i+1].(string))-2] ]; exists {
							log.Info("Substituting deepvariable ", vars[i], " : old value : ", vars[i+1], " for new value: ", renderValue(val).(string))
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

func pipelineIDFunc(vars []varMap, pipelines spinnaker.PipelineAPI) interface{} {
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
		id, err := pipelines.GetPipelineID(app, pipelineName)
		if err != nil {
			log.Errorf("could not get pipeline id for app %s, pipeline %s, err = %v", app, pipelineName, err)
		}
		return id
	}
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
			s, isStr := defaultVal[0].(string)
			if isStr {
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
		log.Errorf("could not download %s/%s/%s", org, repo, path)
		return nil, err
	}

	// Preprocess to stringify any json args in calls to modules.
	contents, err = preprocessor.Preprocess(contents)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Extract global vars if we're processing a dinghyfile (and not a module)
	if filepath.Base(path) == b.DinghyfileName {
		gvs := preprocessor.ParseGlobalVars(contents)
		gvMap, ok := gvs.(map[string]interface{})
		if !ok {
			log.Error("Could not extract global vars")
		} else if len(gvMap) > 0 {
			vars = append(vars, gvMap)
		} else {
			log.Info("No global vars found in dinghyfile")
		}
	}

	funcMap := template.FuncMap{
		"module":     moduleFunc(b, org, deps, vars),
		"appModule":  moduleFunc(b, org, deps, vars),
		"pipelineID": pipelineIDFunc(vars, b.PipelineAPI),
		"var":        varFunc(vars),
	}

	// Parse the downloaded template.
	tmpl, err := template.New("dinghy-render").Funcs(funcMap).Parse(contents)
	if err != nil {
		log.Errorf("template parsing: %s", err)
		return nil, err
	}

	// Run the template to verify the output.
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, "")
	if err != nil {
		log.Errorf("template execution: %s", err)
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
