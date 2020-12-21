package yaml

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/armory/dinghy/pkg/dinghyfile/pipebuilder"
	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/git"
	"github.com/armory/dinghy/pkg/preprocessor"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"text/template"
	"time"
)

type DinghyfileYamlParser struct {
	Builder *dinghyfile.PipelineBuilder
}

func NewDinghyfileYamlParser(b *dinghyfile.PipelineBuilder) *DinghyfileYamlParser {
	return &DinghyfileYamlParser{Builder: b}
}

func (r *DinghyfileYamlParser) SetBuilder(b *dinghyfile.PipelineBuilder) {
	r.Builder = b
}

func (r *DinghyfileYamlParser) parseValue(val interface{}) interface{} {
	var err error
	r.Builder.Logger.Debugf("parsing %v", val)
	if yamlStr, ok := val.(string); ok && len(yamlStr) > 0 {
		if yamlStr[0] == '{' {
			err = yaml.Unmarshal([]byte(yamlStr), &val)
		}

		/*
		 * NOTE: Unlike json, whitespace matters in yaml. If we are to parse arrays in yaml during a value parse we
		 *       end up with the following raw string which doesn't have enough context to be spaced properly:
		 *
		 *  [ "foo", "baz" ] produces:
		 *  - foo
		 *  - baz
		 *
		 *   Once smashed together during the rest of the parse run the resulting strings lose all leading whitespace
		 *   because the entire template isn't read in as yaml all at once.
		 *
		 *   In this instance, it is better to leave json-esque array structures intact as they are still valid yaml.
		 *
		 *  This function is left here for posterity.
		 *
		 *	if yamlStr[0] == '[' {
		 *		err = yaml.Unmarshal([]byte(yamlStr), &val)
		 *	}
		 *
		 */
	}
	if err != nil {
		r.Builder.Logger.Errorf("Error parsing value %s: %s", val.(string), err.Error())
	}

	return val
}

func (r *DinghyfileYamlParser) renderValue(val interface{}) (interface{}, error) {
	// If it's an unserialized YAML array, serialize it back to YAML.
	if newval, ok := val.([]interface{}); ok {
		buf, err := yaml.Marshal(newval)
		if err != nil {
			return "", fmt.Errorf("unable to yaml.marshal array value %v", val)
		}
		return string(buf), nil
	}

	// If it's an unserialized YAML object, serialize it back to YAML.
	if newval, ok := val.(map[string]interface{}); ok {
		buf, err := yaml.Marshal(newval)
		if err != nil {
			return "", fmt.Errorf("unable to yaml.marshal map value %v", val)
		}
		return string(buf), nil
	}

	// Return value as is.
	return val, nil
}

func (r *DinghyfileYamlParser) varFunc(vars []dinghyfile.VarMap) interface{} {
	return func(varName string, defaultVal ...interface{}) (interface{}, error) {
		for _, vm := range vars {
			if val, exists := vm[varName]; exists {
				return r.renderValue(val)
			}
		}

		if len(defaultVal) <= 0 {
			return "", nil
		}

		s, isStr := defaultVal[0].(string)
		if !isStr || len(s) <= 0 {
			return defaultVal[0], nil
		}

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
		return defaultVal[0], nil
	}
}

func (r *DinghyfileYamlParser) pipelineIDFunc(vars []dinghyfile.VarMap) interface{} {
	return func(app, pipelineName string) (string, error) {
		for _, vm := range vars {
			if val, exists := vm["triggerApp"]; exists {
				appr, err := r.renderValue(val)
				if err != nil {
					return "", err
				}
				app = appr.(string)
				r.Builder.Logger.Info("Substituting pipeline triggerApp: ", app)
			}
			if val, exists := vm["triggerPipeline"]; exists {
				pn, err := r.renderValue(val)
				if err != nil {
					return "", err
				}
				pipelineName = pn.(string)
				r.Builder.Logger.Info("Substituting pipeline triggerPipeline: ", pipelineName)
			}
		}
		id, err := r.Builder.GetPipelineByID(app, pipelineName)
		if err != nil {
			return "", fmt.Errorf("could not get pipeline id for app %s, pipeline %s, err = %v", app, pipelineName, err)
		}
		return id, nil
	}
}

func (r *DinghyfileYamlParser) moduleFunc(org string, repo string, branch string, deps map[string]bool, allVars []dinghyfile.VarMap) interface{} {
	return func(mod string, vars ...interface{}) (string, error) {
		return moduleFunction(org, mod, r, repo, branch, deps, vars, allVars)
	}
}

func moduleFunction(org string, mod string, r *DinghyfileYamlParser, repo string, branch string, deps map[string]bool, vars []interface{}, allVars []dinghyfile.VarMap) (string, error) {

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
	newVars := make(dinghyfile.VarMap)
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
						newValue, err := r.renderValue(val)
						if err != nil {
							return "", err
						}
						r.Builder.Logger.Info("Substituting deepvariable ", vars[i], " : old value : ", deepVariable, " for new value: ", newValue.(string))
						vars[i+1] = r.parseValue(val)
					}
				}
			}
		}
		newVars[key] = r.parseValue(vars[i+1])
	}

	result, err := r.Parse(org, repo, mod, branch, append([]dinghyfile.VarMap{newVars}, allVars...))
	if err != nil {
		r.Builder.Logger.Errorf("Error rendering imported module '%s': %s", mod, err.Error())
	}
	return result.String(), nil
}

func (r *DinghyfileYamlParser) makeSlice(args ...interface{}) []interface{} {
	return args
}

func (r *DinghyfileYamlParser) Parse(org, repo, path, branch string, vars []dinghyfile.VarMap) (*bytes.Buffer, error) {
	module := true
	event := &events.Event{
		Start: time.Now().UTC().Unix(),
		Org:   org,
		Repo:  repo,
		Path:  path,
	}

	gitInfo := git.GitInfo{
		RawData: r.Builder.PushRaw,
		Org: org,
		Repo: repo,
		Path: path,
		Branch: branch,
	}

	deps := make(map[string]bool)

	// Download the template being parsed.
	r.Builder.Logger.Info("Downloading ", path)
	contents, err := r.Builder.Downloader.Download(org, repo, path, branch)
	if err != nil {
		r.Builder.Logger.Error("Failed to download")
		return nil, err
	}

	// Preprocess to stringify any yaml args in calls to modules.
	//NOTE: Uncertain this is necessary in a yaml context
	r.Builder.Logger.Info("Preprocessing ", path)
	contents, err = preprocessor.Preprocess(contents)
	if err != nil {
		r.Builder.Logger.Error("Failed to preprocess")
		return nil, err
	}

	// Extract global vars if we're processing a dinghyfile (and not a module)
	if filepath.Base(path) == r.Builder.DinghyfileName {
		//module = false
		gvs, err := ParseGlobalVars(contents, gitInfo)
		if err != nil {
			if gvs == nil {
				r.Builder.Logger.Errorf("Failed to parse global vars: %s", err.Error())
				return nil, err
			}
			r.Builder.Logger.Warnf("Failed to parse global vars, but continuing anyway: %s", err.Error())
		}

		gvMap, ok := gvs.(map[string]interface{})
		if !ok {
			return nil, errors.New("Could not extract global vars")
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
	if r.Builder.Action == pipebuilder.Validate {
		moduleBranch = "master"
	}

	funcMap := template.FuncMap{
		"module":        r.moduleFunc(r.Builder.TemplateOrg, r.Builder.TemplateRepo, moduleBranch, deps, vars),
		"local_module":  r.localModuleFunc(org, repo, branch, deps, vars),
		"appModule":     r.moduleFunc(r.Builder.TemplateOrg, r.Builder.TemplateRepo, moduleBranch, deps, vars),
		"pipelineID":    r.pipelineIDFunc(vars),
		"var":           r.varFunc(vars),
		"makeSlice":     r.makeSlice,
	}

	// Parse the downloaded template.
	tmpl, err := template.New("dinghy-render").Funcs(sprig.TxtFuncMap()).Funcs(funcMap).Parse(contents)
	if err != nil {
		r.Builder.Logger.Error("Failed to parse template")
		return nil, err
	}

	// Run the template to verify the output.
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, gitInfo)
	if err != nil {
		r.Builder.Logger.Error("Failed to execute buffer")
		return nil, err
	}

	// Record the dependencies we ran into.
	depUrls := make([]string, 0)
	for dep := range deps {
		depUrls = append(depUrls, dep)
	}
	r.Builder.Depman.SetDeps(r.Builder.Downloader.EncodeURL(org, repo, path, branch), depUrls)

	event.End = time.Now().UTC().Unix()
	eventType := "render"
	event.Dinghyfile = buf.String()
	event.Module = module
	r.Builder.EventClient.SendEvent(eventType, event)

	return buf, nil
}


func (r *DinghyfileYamlParser) localModuleFunc(org string, repo string, branch string, deps map[string]bool, allVars []dinghyfile.VarMap) interface{} {
	return func(mod string, vars ...interface{}) (string, error) {
		if r.Builder.TemplateOrg == org && r.Builder.TemplateRepo == repo {
			return "", fmt.Errorf("%v is a local_module, calling local_module from a module is not allowed", mod)
		} else {
			return moduleFunction(org, mod, r, repo, branch, deps, vars, allVars)
		}
	}
}
