package dinghyfile

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/armory-io/dinghy/pkg/spinnaker"
)

func TestUpdateDinghyfile(t *testing.T) {

	cases := map[string]struct {
		dinghyfile []byte
		spec       spinnaker.ApplicationSpec
	}{
		"no_spec": {
			dinghyfile: []byte(`{
				"application": "application"
			}`),
			spec: spinnaker.ApplicationSpec{
				Name:  "application",
				Email: DefaultEmail,
				DataSources: spinnaker.DataSourcesSpec{
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
			spec: spinnaker.ApplicationSpec{
				Name:  "specname",
				Email: DefaultEmail,
				DataSources: spinnaker.DataSourcesSpec{
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
			spec: spinnaker.ApplicationSpec{
				Name:  "specname",
				Email: "somebody@email.com",
				DataSources: spinnaker.DataSourcesSpec{
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
			spec: spinnaker.ApplicationSpec{
				Name:  "specname",
				Email: "somebody@email.com",
				DataSources: spinnaker.DataSourcesSpec{
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
