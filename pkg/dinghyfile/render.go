package dinghyfile

import (
	"bytes"
	"encoding/json"
	"errors"

	"text/template"

	"github.com/armory-io/dinghy/pkg/preprocessor"
	"github.com/armory-io/dinghy/pkg/settings"
	log "github.com/sirupsen/logrus"
)

// Render renders the template
func (b PipelineBuilder) Render(org, repo, path string, v []interface{}) *bytes.Buffer {
	deps := make(map[string]bool)

	funcMap := template.FuncMap{
		"module": func(mod string, vars ...interface{}) string {
			var tmp map[string]interface{}
			rendered := b.Render(settings.S.TemplateOrg, settings.S.TemplateRepo, mod, vars)
			err := json.Unmarshal(rendered.Bytes(), &tmp)
			if err != nil {
				log.Fatal("could not unmarshal module after rendering: ", mod, " err: ", err)
			}
			allVars := append(vars, v...)

			if len(vars)%2 != 0 {
				log.Fatal(errors.New("invalid number of args to module: " + mod))
			}

			for i := 0; i < len(allVars); i += 2 {
				key, ok := allVars[i].(string)
				if !ok {
					log.Fatal(errors.New("dict keys must be strings in module: " + mod))
				}

				val, exists := tmp[key]
				if !exists {
					continue
				}

				newVal := vars[i+1]
				log.Info("newval: ", newVal)

				if jsonStr, ok := newVal.(string); ok {
					if jsonStr[0] == '{' {
						json.Unmarshal([]byte(jsonStr), &newVal)
					}
					if jsonStr[0] == '[' {
						json.Unmarshal([]byte(jsonStr), &newVal)
					}
				}

				tmp[key] = newVal
				log.Info(" ** variable substitution in ", mod, " for key: ", key, ", value ", val, " --> ", tmp[key])
			}

			byt, err := json.Marshal(tmp)
			if err != nil {
				log.Fatal("could not marshal variable substituted json for module: ", mod, err)
			}

			child := b.downloader.EncodeURL(org, settings.S.TemplateRepo, mod)
			if _, exists := deps[child]; !exists {
				deps[child] = true
			}

			return string(byt)
		},
	}

	contents, err := b.downloader.Download(org, repo, path)
	if err != nil {
		log.Fatalf("could not download %s/%s/%s", org, repo, path)
	}

	// Preprocess to stringify any json args in calls to modules.
	contents = preprocessor.Preprocess(contents)

	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("moduleTest").Funcs(funcMap).Parse(contents)
	if err != nil {
		log.Fatalf("template parsing: %s", err)
	}

	// Run the template to verify the output.
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, "")
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}

	depUrls := make([]string, 0)
	for dep := range deps {
		depUrls = append(depUrls, dep)
	}

	b.depman.SetDeps(b.downloader.EncodeURL(org, repo, path), depUrls)
	return buf
}
