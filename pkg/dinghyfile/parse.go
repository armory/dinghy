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
	"fmt"
	"github.com/armory/dinghy/pkg/dinghyfile/pipebuilder"
	"github.com/armory/dinghy/pkg/git"
	"path/filepath"
	"time"

	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/preprocessor"
)

type DinghyfileParser struct {
	Builder *PipelineBuilder
}

func NewDinghyfileParser(b *PipelineBuilder) *DinghyfileParser {
	return &DinghyfileParser{Builder: b}
}

func (r *DinghyfileParser) SetBuilder(b *PipelineBuilder) {
	r.Builder = b
}

func (r *DinghyfileParser) parseValue(val interface{}) interface{} {
	var err error
	if jsonStr, ok := val.(string); ok && len(jsonStr) > 0 {
		if jsonStr[0] == '{' {
			err = json.Unmarshal([]byte(jsonStr), &val)
		}
		if jsonStr[0] == '[' {
			err = json.Unmarshal([]byte(jsonStr), &val)
		}
	}
	if err != nil {
		r.Builder.Logger.Errorf("Error parsing value %s: %s", val.(string), err.Error())
	}

	return val
}

// TODO: this function errors, it should be returning the error to the caller to be handled
func (r *DinghyfileParser) moduleFunc(org string, repo string, branch string, deps map[string]bool, allVars []VarMap) interface{} {
	return func(mod string, vars ...interface{}) (string, error) {
		return moduleFunction(org, mod, r, repo, branch, deps, vars, allVars)
	}
}

func moduleFunction(org string, mod string, r *DinghyfileParser, repo string, branch string, deps map[string]bool, vars []interface{}, allVars []VarMap) (string, error) {
	// Don't bother if the TemplateOrg isn't set.
	if org == "" {
		return "", fmt.Errorf("Cannot load module %s; templateOrg not configured", mod)
	}

	// Record the dependency.
	child := r.Builder.Downloader.EncodeURL(org, repo, mod, branch)
	if _, exists := deps[child]; !exists {
		deps[child] = true
	}

	length := len(vars)
	if length%2 != 0 {
		r.Builder.Logger.Warnf("odd number of parameters received to module %s", mod)
	}

	// Convert module argument pairs to key/value map
	newVars := make(VarMap)
	for i := 0; i+1 < length; i += 2 {
		key, ok := vars[i].(string)
		if !ok {
			r.Builder.Logger.Errorf("dict keys must be strings in module: %s", mod)
			return "", fmt.Errorf("dict keys must be strings in module: %s", mod)
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

	result, err := r.Parse(org, repo, mod, branch, append([]VarMap{newVars}, allVars...))
	if err != nil {
		r.Builder.Logger.Errorf("error rendering imported module '%s': %s", mod, err.Error())
		return "", fmt.Errorf("error rendering imported module '%s': %s", mod, err.Error())
	}
	return result.String(), nil
}

// TODO: this function errors, it should be returning the error to the caller to be handled
func (r *DinghyfileParser) pipelineIDFunc(vars []VarMap) interface{} {
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
func (r *DinghyfileParser) renderValue(val interface{}) interface{} {
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

func (r *DinghyfileParser) varFunc(vars []VarMap) interface{} {
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

func (r *DinghyfileParser) makeSlice(args ...interface{}) []interface{} {
	return args
}

// Parse parses the template
func (r *DinghyfileParser) Parse(org, repo, path, branch string, vars []VarMap) (*bytes.Buffer, error) {
	module := true
	event := &events.Event{
		Start:  time.Now().UTC().Unix(),
		Org:    org,
		Repo:   repo,
		Path:   path,
		Branch: branch,
		End:    time.Now().UTC().Unix(),
	}

	gitInfo := git.GitInfo{
		r.Builder.PushRaw,
		org, repo, path, branch,
	}

	deps := make(map[string]bool)

	// Download the template being parsed.
	contents, err := r.Builder.Downloader.Download(org, repo, path, branch)
	if err != nil {
		r.Builder.Logger.Errorf("Failed to download %s/%s/%s/%s", org, repo, path, branch)
		// we don't actually have a dinghyfile we can send at this point
		r.Builder.EventClient.SendEvent("parse-err-download", event)
		return nil, err
	}

	// Preprocess to stringify any json args in calls to modules.
	contents, err = preprocessor.Preprocess(contents)
	if err != nil {
		r.Builder.Logger.Errorf("Failed to preprocess:\n %s", contents)
		event.Dinghyfile = contents
		r.Builder.EventClient.SendEvent("parse-err-preprocess", event)
		return nil, err
	}

	// Validate if module is parsed correctly
	if r.Builder.Action == pipebuilder.Validate && repo == r.Builder.TemplateRepo {
		err = preprocessor.ContentShouldBeParsedCorrectly(contents)
		if err != nil {
			r.Builder.Logger.Errorf("Failed to parse module:\n %s", contents)
			r.Builder.EventClient.SendEvent("parse-err-module", event)
			return nil, err
		}
	}

	// Extract global vars if we're processing a dinghyfile (and not a module)
	isDinghyfile := filepath.Base(path) == r.Builder.DinghyfileName
	if isDinghyfile {
		module = false
		gvs, err := preprocessor.ParseGlobalVars(contents, gitInfo)
		if err != nil {
			r.Builder.Logger.Errorf("Failed to parse global vars:\n %s", contents)
			event.Dinghyfile = contents
			r.Builder.EventClient.SendEvent("parse-err-globalvar", event)
			return nil, err
		}

		gvMap, ok := gvs.(map[string]interface{})
		if !ok {
			event.Dinghyfile = contents
			r.Builder.EventClient.SendEvent("parse-err-globalvar", event)
			return nil, fmt.Errorf("could not extract global vars from:\n %s", contents)
		} else if len(gvMap) > 0 {
			vars = append(vars, gvMap)
		} else {
			r.Builder.Logger.Info("No global vars found in dinghyfile")
		}
		r.Builder.GlobalVariablesMap = gvMap
	}

	// If we are validating then check always against the modules in master since current branch will
	// not exists in templare repo
	var moduleBranch = branch
	// if we are doing a update on template repo, we should test against the branch
	if r.Builder.Action == pipebuilder.Validate && r.Builder.TemplateRepo != repo {
		moduleBranch = "master"
	}
	// NOTE:  I don't think moduleFunc needs to take branch argument;
	// moduleFunc should be able to figure out the branch needed from the
	// configuration (since it has to have access to TemplateOrg and TemplateRepo
	// anyway.  But MAYBE we want to actually figure that out here where we
	// have an application in context?  So for now, hardcoding module branch
	// to "master"
	funcMap := template.FuncMap{
		"module":       r.moduleFunc(r.Builder.TemplateOrg, r.Builder.TemplateRepo, moduleBranch, deps, vars),
		"local_module": r.localModuleFunc(org, repo, branch, isDinghyfile, deps, vars),
		"appModule":    r.moduleFunc(r.Builder.TemplateOrg, r.Builder.TemplateRepo, moduleBranch, deps, vars),
		"pipelineID":   r.pipelineIDFunc(vars),
		"var":          r.varFunc(vars),
		"makeSlice":    r.makeSlice,
	}

	// Parse the downloaded template.
	tmpl, err := template.New("dinghy-render").Funcs(sprig.TxtFuncMap()).Funcs(funcMap).Parse(contents)
	if err != nil {
		r.Builder.Logger.Errorf("Failed to parse template:\n %s", contents)
		event.Dinghyfile = contents
		r.Builder.EventClient.SendEvent("parse-err-gotemplate-funcs", event)
		return nil, err
	}

	// Run the template to verify the output.
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, gitInfo)
	if err != nil {
		r.Builder.Logger.Errorf("Failed to execute buffer:\n %s\nError: %s", contents, err.Error())
		event.Dinghyfile = contents
		r.Builder.EventClient.SendEvent("parse-err-bytebuffer", event)
		return nil, err
	}

	// Record the dependencies we ran into.
	depUrls := make([]string, 0)
	for dep := range deps {
		depUrls = append(depUrls, dep)
	}
	r.Builder.Depman.SetDeps(r.Builder.Downloader.EncodeURL(org, repo, path, branch), depUrls)
	if filepath.Base(path) == r.Builder.DinghyfileName && !r.Builder.RebuildingModules {
		result, errRaw := json.Marshal(r.Builder.PushRaw)
		if errRaw != nil {
			r.Builder.Logger.Errorf("Failed to parse rawdata:\n %s", r.Builder.PushRaw)
			r.Builder.EventClient.SendEvent("parse-err-rawdata", event)
			return nil, errRaw
		}
		r.Builder.Depman.SetRawData(r.Builder.Downloader.EncodeURL(org, repo, path, branch), string(result))
	}

	event.Dinghyfile = buf.String()
	event.Module = module
	r.Builder.EventClient.SendEvent("parse", event)

	return buf, nil
}

func (r *DinghyfileParser) localModuleFunc(org string, repo string, branch string, isDinghyfile bool, deps map[string]bool, allVars []VarMap) interface{} {
	return func(mod string, vars ...interface{}) (string, error) {
		if r.Builder.TemplateOrg == org && r.Builder.TemplateRepo == repo && !isDinghyfile {
			return "", fmt.Errorf("%v is a local_module, calling local_module from a module is not allowed", mod)
		} else {
			return moduleFunction(org, mod, r, repo, branch, deps, vars, allVars)
		}
	}
}
