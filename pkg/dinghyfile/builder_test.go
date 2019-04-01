package dinghyfile

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/armory/plank"
)

func TestUpdateDinghyfile(t *testing.T) {

	cases := map[string]struct {
		dinghyfile []byte
		spec       plank.Application
	}{
		"no_spec": {
			dinghyfile: []byte(`{
				"application": "application"
			}`),
			spec: plank.Application{
				Name:  "application",
				Email: DefaultEmail,
				DataSources: plank.DataSourcesType{
					Enabled:  []string{},
					Disabled: []string{},
				},
			},
		},
		"spec": {
			dinghyfile: []byte(`{
				"application": "application",
				"spec": {
					"name": "specname"
				}
			}`),
			spec: plank.Application{
				Name:  "specname",
				Email: DefaultEmail,
				DataSources: plank.DataSourcesType{
					Enabled:  []string{},
					Disabled: []string{},
				},
			},
		},
		"spec_with_email": {
			dinghyfile: []byte(`{
				"application": "application",
				"spec": {
					"name": "specname",
					"email": "somebody@email.com"
				}
			}`),
			spec: plank.Application{
				Name:  "specname",
				Email: "somebody@email.com",
				DataSources: plank.DataSourcesType{
					Enabled:  []string{},
					Disabled: []string{},
				},
			},
		},
		"spec_with_canary": {
			dinghyfile: []byte(`{
				"application": "application",
				"spec": {
					"name": "specname",
					"email": "somebody@email.com",
					"dataSources": {
						"disabled":[],
						"enabled":["canaryConfigs"]
					}
				}
			}`),
			spec: plank.Application{
				Name:  "specname",
				Email: "somebody@email.com",
				DataSources: plank.DataSourcesType{
					Enabled:  []string{"canaryConfigs"},
					Disabled: []string{},
				},
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			d, _ := UpdateDinghyfile(c.dinghyfile)
			assert.Equal(t, d.ApplicationSpec, c.spec)
		})
	}
}
