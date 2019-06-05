/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package dinghyfile

/*
 * NOTE:  This is actually the Dinghyfile renderer; it should probably be
 * renamed accordingly.
 */

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"text/template"

	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/preprocessor"
)

type DinghyfileRenderer struct {
	Builder *PipelineBuilder
}

func NewDinghyfileRenderer(b *PipelineBuilder) *DinghyfileRenderer {
	return &DinghyfileRenderer{Builder: b}
}

func (r *DinghyfileRenderer) parseValue(val interface{}) interface{} {
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

// TODO: this function errors, it should be returning the error to the caller to be handled
func (r *DinghyfileRenderer) moduleFunc(org string, deps map[string]bool, allVars []varMap) interface{} {
	return func(mod string, vars ...interface{}) string {
		// Record the dependency.
		child := r.Builder.Downloader.EncodeURL(org, r.Builder.TemplateRepo, mod)
		if _, exists := deps[child]; !exists {
			deps[child] = true
		}

		length := len(vars)
		if length%2 != 0 {
			r.Builder.Logger.Warnf("odd number of parameters received to module %s", mod)
		}

		// Convert module argument pairs to key/value map
		newVars := make(varMap)
		for i := 0; i+1 < length; i += 2 {
			key, ok := vars[i].(string)
			if !ok {
				r.Builder.Logger.Errorf("dict keys must be strings in module: %s", mod)
				return ""
			}

			// checks for deepvariables, passes all the way down values from dinghyFile to module inside module
			deepVariable, ok := vars[i+1].(string)
			if ok && len(deepVariable) > 6 {
				if deepVariable[0:5] == "{{var" {
					for _, vm := range allVars {
						if val, exists := vm[deepVariable[6:len(deepVariable)-2]]; exists {
							r.Builder.Logger.Info("Substituting deepvariable ", vars[i], " : old value : ", deepVariable, " for new value: ", r.renderValue(val).(string))
							vars[i+1] = r.parseValue(val)
						}
					}
				}
			}
			newVars[key] = r.parseValue(vars[i+1])
		}

		result, _ := r.Render(r.Builder.TemplateOrg, r.Builder.TemplateRepo, mod, append([]varMap{newVars}, allVars...))
		return result.String()
	}
}

// TODO: this function errors, it should be returning the error to the caller to be handled
func (r *DinghyfileRenderer) pipelineIDFunc(vars []varMap) interface{} {
	return func(app, pipelineName string) string {
		for _, vm := range vars {
			if val, exists := vm["triggerApp"]; exists {
				app = r.renderValue(val).(string)
				r.Builder.Logger.Info("Substituting pipeline triggerApp: ", app)
			}
			if val, exists := vm["triggerPipeline"]; exists {
				pipelineName = r.renderValue(val).(string)
				r.Builder.Logger.Info("Substituting pipeline triggerPipeline: ", pipelineName)
			}
		}
		id, err := r.Builder.GetPipelineByID(app, pipelineName)
		if err != nil {
			r.Builder.Logger.Errorf("could not get pipeline id for app %s, pipeline %s, err = %v", app, pipelineName, err)
		}
		return id
	}
}

// TODO: this function errors, it should be returning the error to the caller to be handled
func (r *DinghyfileRenderer) renderValue(val interface{}) interface{} {
	// If it's an unserialized JSON array, serialize it back to JSON.
	if newval, ok := val.([]interface{}); ok {
		buf, err := json.Marshal(newval)
		if err != nil {
			r.Builder.Logger.Errorf("unable to json.marshal array value %v", val)
			return ""
		}
		return string(buf)
	}

	// If it's an unserialized JSON object, serialize it back to JSON.
	if newval, ok := val.(map[string]interface{}); ok {
		buf, err := json.Marshal(newval)
		if err != nil {
			r.Builder.Logger.Errorf("unable to json.marshal map value %v", val)
			return ""
		}
		return string(buf)
	}

	// Return value as is.
	return val
}

func (r *DinghyfileRenderer) varFunc(vars []varMap) interface{} {
	return func(varName string, defaultVal ...interface{}) interface{} {
		for _, vm := range vars {
			if val, exists := vm[varName]; exists {
				return r.renderValue(val)
			}
		}

		if len(defaultVal) > 0 {
			s, isStr := defaultVal[0].(string)

			if isStr && len(s) > 0 {
				if s[0] == '@' {
					// handle the case where the default value is another variable
					// see ENG-1921 for use case. e.g.,:
					// {{ var "servicename" ?: "@application" }}

					nested := s[1:]
					for _, vm := range vars {
						if val, exists := vm[nested]; exists {
							r.Builder.Logger.Info("Substituting nested variable: ", nested, ", val: ", val)
							return r.renderValue(val)
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
func (r *DinghyfileRenderer) Render(org, repo, path string, vars []varMap) (*bytes.Buffer, error) {
	module := true
	event := &events.Event{
		Start: time.Now().UTC().Unix(),
		Org:   org,
		Repo:  repo,
		Path:  path,
	}

	deps := make(map[string]bool)

	// Download the template being rendered.
	contents, err := r.Builder.Downloader.Download(org, repo, path)
	if err != nil {
		r.Builder.Logger.Error("Failed to download")
		return nil, err
	}

	// Preprocess to stringify any json args in calls to modules.
	contents, err = preprocessor.Preprocess(contents)
	if err != nil {
		r.Builder.Logger.Error("Failed to preprocess")
		return nil, err
	}

	// Extract global vars if we're processing a dinghyfile (and not a module)
	if filepath.Base(path) == r.Builder.DinghyfileName {
		module = false
		gvs, err := preprocessor.ParseGlobalVars(contents)
		if err != nil {
			r.Builder.Logger.Error("Failed to parse global vars")
			return nil, err
		}

		gvMap, ok := gvs.(map[string]interface{})
		if !ok {
			return nil, errors.New("Could not extract global vars")
		} else if len(gvMap) > 0 {
			vars = append(vars, gvMap)
		} else {
			r.Builder.Logger.Info("No global vars found in dinghyfile")
		}
	}

	funcMap := template.FuncMap{
		"module":     r.moduleFunc(org, deps, vars),
		"appModule":  r.moduleFunc(org, deps, vars),
		"pipelineID": r.pipelineIDFunc(vars),
		"var":        r.varFunc(vars),
	}

	// Parse the downloaded template.
	tmpl, err := template.New("dinghy-render").Funcs(funcMap).Parse(contents)
	if err != nil {
		r.Builder.Logger.Error("Failed to parse template")
		return nil, err
	}

	// Run the template to verify the output.
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, "")
	if err != nil {
		r.Builder.Logger.Error("Failed to execute buffer")
		return nil, err
	}

	// Record the dependencies we ran into.
	depUrls := make([]string, 0)
	for dep := range deps {
		depUrls = append(depUrls, dep)
	}
	r.Builder.Depman.SetDeps(r.Builder.Downloader.EncodeURL(org, repo, path), depUrls)

	event.End = time.Now().UTC().Unix()
	eventType := "render"
	event.Dinghyfile = buf.String()
	event.Module = module
	r.Builder.EventClient.SendEvent(eventType, event)

	return buf, nil
}
