package dinghyfile

import (
	"bytes"
	"encoding/json"
	"errors"

	"text/template"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/preprocessor"
	"github.com/armory-io/dinghy/pkg/settings"
	log "github.com/sirupsen/logrus"
)

// Render function renders the template
func Render(c cache.CacheStore, fileName, file, gitOrg, gitRepo string, f git.Downloader, v []interface{}) *bytes.Buffer {
	deps := make([]string, 0)

	funcMap := template.FuncMap{
		"module": func(mod string, vars ...interface{}) string {
			// this is the temp struct we decode the module into for
			// variable substitution
			var tmp map[string]interface{}

			dat, err := f.Download(settings.S.TemplateOrg, settings.S.TemplateRepo, mod)
			if err != nil {
				log.Fatal("could not read module: ", mod, " err: ", err)
			}

			rendered := Render(c, mod, dat, settings.S.TemplateOrg, settings.S.TemplateRepo, f, vars)
			err = json.Unmarshal(rendered.Bytes(), &tmp)
			if err != nil {
				log.Fatal("could not unmarshal module after rendering: ", mod, " err: ", err)
			}
			allVars := append(vars, v...)
			if len(allVars)%2 != 0 {
				log.Fatal(errors.New("invalid number of args to module: " + mod))
			}

			for i := 0; i < len(allVars); i += 2 {
				key, ok := allVars[i].(string)
				if !ok {
					log.Fatal(errors.New("dict keys must be strings in module: " + mod))
				}
				if val, ok := tmp[key]; ok {
					newVal := allVars[i+1]
					log.Info("newval: ", newVal)
					if jsonStr, ok := newVal.(string); ok {
						/* act on str */
						var js map[string]interface{}
						var ar interface{}
						if jsonStr[0] == '{' {
							json.Unmarshal([]byte(jsonStr), &js)
							tmp[key] = js
						} else if jsonStr[0] == '[' {
							json.Unmarshal([]byte(jsonStr), &ar)
							tmp[key] = ar
						} else {
							/* not json object */
							tmp[key] = newVal
						}
					} else {
						/* not string */
						tmp[key] = newVal
					}
					log.Info(" ** variable substitution in ", mod, " for key: ", key, ", value ", val, " --> ", tmp[key])
				}
			}
			byt, err := json.Marshal(tmp)
			if err != nil {
				log.Fatal("could not marshal variable substituted json for module: ", mod, err)
			}

			child := f.GitURL(gitOrg, settings.S.TemplateRepo, mod)
			found := false
			for _, dep := range deps {
				if dep == child {
					found = true
					break
				}
			}
			if !found {
				deps = append(deps, child)
			}

			return string(byt)
		},
	}

	// preprocess to stringify any json args in calls to modules
	file = preprocessor.Preprocess(file)

	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("moduleTest").Funcs(funcMap).Parse(file)
	if err != nil {
		log.Fatalf("template parsing: %s", err)
	}

	// Run the template to verify the output.
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, "")
	if err != nil {
		log.Fatalf("template execution: %s", err)
	}

	c.SetDeps(f.GitURL(gitOrg, gitRepo, fileName), deps...)
	return buf
}
