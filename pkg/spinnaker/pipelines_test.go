package spinnaker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//delete pipelines
//get pipelines
//get piipeliens by apli
//create pipeline

type fakeFront50 struct {
	Front50API
	getPipelinesForApplicationResponse []byte
	getPipelinesForApplicationError    error
	createPipelineError                error
	applicationExistsResponse          bool
	updatePipelineError                error
	newApplicationError                error
	deletePipelineError                error
}

func (fakeFront50 *fakeFront50) GetPipelinesForApplication(appName string) ([]byte, error) {
	return fakeFront50.getPipelinesForApplicationResponse, fakeFront50.getPipelinesForApplicationError
}

func (fakeFront50 *fakeFront50) CreatePipeline(p Pipeline) error {
	return fakeFront50.createPipelineError
}

func (fakeFront50 *fakeFront50) ApplicationExists(appName string) bool {
	return fakeFront50.applicationExistsResponse
}

func (fakeFront50 *fakeFront50) NewApplication(spec ApplicationSpec) error {
	return fakeFront50.newApplicationError
}

func (fakeFront50 *fakeFront50) UpdatePipeline(id string, p Pipeline) error {
	return fakeFront50.updatePipelineError
}

func (fakeFront50 *fakeFront50) DeletePipeline(appName, pipelineName string) error {
	return fakeFront50.deletePipelineError
}

func TestDefaultPipelineAPI_GetPipelineID(t *testing.T) {
	cases := map[string]struct {
		expected     string
		fakeFront50  *fakeFront50
		expectsError bool
	}{
		"happy path": {
			expected:     "12345",
			expectsError: false,
			fakeFront50: &fakeFront50{
				getPipelinesForApplicationResponse: []byte(`[{"id": "12345", "name": "test"}]`),
				getPipelinesForApplicationError:    nil,
				createPipelineError:                nil,
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			pipelineAPI := DefaultPipelineAPI{Front50API: c.fakeFront50}
			id, err := pipelineAPI.GetPipelineID("test", "test")
			if c.expectsError {
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, c.expected, id)
		})
	}
}

func TestDefaultPipelineAPI_UpdatePipelines(t *testing.T) {
	cases := map[string]struct {
		fakeFront50  *fakeFront50
		expectsError bool
		spec         ApplicationSpec
		p            []Pipeline
		config       UpdatePipelineConfig
	}{
		"application already exists": {
			expectsError: false,
			spec:         ApplicationSpec{Name: "test"},
			p: []Pipeline{
				Pipeline{"name": "test", "application": "test"},
			},
			config: UpdatePipelineConfig{},
			fakeFront50: &fakeFront50{
				applicationExistsResponse: true,
				updatePipelineError:       nil,
			},
		},
		"application needs create": {
			expectsError: false,
			spec:         ApplicationSpec{Name: "test"},
			p: []Pipeline{
				Pipeline{"name": "test", "application": "test"},
			},
			config: UpdatePipelineConfig{},
			fakeFront50: &fakeFront50{
				applicationExistsResponse:          false,
				updatePipelineError:                nil,
				newApplicationError:                nil,
				getPipelinesForApplicationResponse: []byte(`[{"id": "12345", "name": "test"}]`),
				getPipelinesForApplicationError:    nil,
			},
		},
		"update with delete pipelines": {
			expectsError: false,
			spec:         ApplicationSpec{Name: "test"},
			p: []Pipeline{
				Pipeline{"name": "test", "application": "test"},
				//Pipeline{"name": "test-stale", "application": "test"},
			},
			config: UpdatePipelineConfig{DeleteStale: true},
			fakeFront50: &fakeFront50{
				applicationExistsResponse:          false,
				updatePipelineError:                nil,
				newApplicationError:                nil,
				getPipelinesForApplicationResponse: []byte(`[{"id": "12345", "name": "test"}, {"id": "789", "name": "test-stale"}]`),
				getPipelinesForApplicationError:    nil,
				deletePipelineError:                nil,
			},
		},
	}

	for testName, c := range cases {
		t.Run(testName, func(t *testing.T) {
			api := DefaultPipelineAPI{
				Front50API: c.fakeFront50,
			}

			err := api.UpdatePipelines(c.spec, c.p, c.config)
			if c.expectsError {
				assert.NotNil(t, err)
			}
		})
	}
}
