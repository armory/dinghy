package dinghyfile

import (
	"bytes"
	"encoding/json"
	"errors"

	"text/template"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/settings"
	log "github.com/sirupsen/logrus"
)

// Render function renders the template
func Render(c cache.Cache, fileName, file, gitOrg, gitRepo string, f git.Downloader) *bytes.Buffer {
	// this is the temp struct we decode the module into for
	// variable substitution
	var tmp map[string]interface{}

	funcMap := template.FuncMap{
		"module": func(mod string, vars ...interface{}) string {
			dat, err := f.Download(settings.S.TemplateOrg, settings.S.TemplateRepo, mod)
			if err != nil {
				log.Fatal("could not read module: ", mod, err)
			}

			rendered := Render(c, mod, dat, settings.S.TemplateOrg, settings.S.TemplateRepo, f)
			json.Unmarshal(rendered.Bytes(), &tmp)

			if len(vars)%2 != 0 {
				log.Fatal(errors.New("invalid number of args to module: " + mod))
			}

			for i := 0; i < len(vars); i += 2 {
				key, ok := vars[i].(string)
				if !ok {
					log.Fatal(errors.New("dict keys must be strings in module: " + mod))
				}
				if val, ok := tmp[key]; ok {
					newVal := vars[i+1]
					log.Info(" ** variable substitution in ", mod, " for key: ", key, ", value ", val, " --> ", newVal)
					tmp[key] = newVal
				}
			}
			byt, err := json.Marshal(tmp)
			if err != nil {
				log.Fatal("could not marshal variable substituted json for module: ", mod, err)
			}
			parent := f.GitURL(gitOrg, gitRepo, fileName)
			child := f.GitURL(gitOrg, settings.S.TemplateRepo, mod)
			c.Add(parent, child)
			return string(byt)
		},
	}

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

	return buf
}
