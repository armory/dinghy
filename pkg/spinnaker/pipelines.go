package spinnaker

import (
	"bytes"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/util"
)

// Pipeline is the structure used by spinnaker
type Pipeline map[string]interface{}

// Lock is embetted in the pipeline if it should be disabled in the UI.
type Lock struct {
	UI            bool `json:"ui"`
	AllowUnlockUI bool `json:"allowUnlockUi"`
}

// Lock disables the pipeline from being edited from the Spinnaker UI.
func (p Pipeline) Lock() {
	p["locked"] = Lock{UI: true, AllowUnlockUI: true}
}

// Name returns the name of the pipeline.
func (p Pipeline) Name() string {
	val, exists := p["name"]
	if exists {
		name := val.(string)
		return name
	}
	return ""
}

// Application returns the app name of the pipeline.
func (p Pipeline) Application() string {
	val, exists := p["application"]
	if exists {
		application := val.(string)
		return application
	}
	return ""
}

// UpdatePipelines posts pipelines to Spinnaker.
type UpdatePipelineConfig struct {
	DeleteStale       bool
	AutolockPipelines string
}

type PipelineAPI interface {
	GetPipelineID(app, pipelineName string) (string, error)
	UpdatePipelines(spec ApplicationSpec, p []Pipeline, config UpdatePipelineConfig) error
}

type DefaultPipelineAPI struct {
	Front50API Front50API
}

func (api *DefaultPipelineAPI) UpdatePipelines(spec ApplicationSpec, p []Pipeline, config UpdatePipelineConfig) (err error) {
	// We currently only use spec if the app is new: TODO update app if the spec changes?
	app := spec.Name
	if !api.Front50API.ApplicationExists(app) {
		api.Front50API.NewApplication(spec)
	}
	ids, _ := api.PipelineIDs(app)
	checklist := make(map[string]bool)
	idToName := make(map[string]string)
	for name, id := range ids {
		checklist[id] = false
		idToName[id] = name
	}
	log.Info("Found pipelines for ", app, ": ", ids)
	for _, pipeline := range p {
		// Add ids to existing pipelines
		if id, exists := ids[pipeline.Name()]; exists {
			log.Debug("Added id ", id, " to pipeline ", pipeline.Name())
			pipeline["id"] = id
			checklist[id] = true
		}
		log.Info("Updating pipeline: " + pipeline.Name())
		if config.AutolockPipelines == "true" {
			log.Debug("Locking pipeline ", pipeline.Name())
			pipeline.Lock()
		}
		err := api.updatePipeline(pipeline)
		if err != nil {
			return err
		}
	}
	if config.DeleteStale {
		// clear existing pipelines that weren't updated
		for id, updated := range checklist {
			if !updated {
				api.Front50API.DeletePipeline(app, idToName[id])
			}
		}
	}
	return err
}

func (api *DefaultPipelineAPI) createEmptyPipeline(app, pipelineName string) error {
	p := Pipeline{
		"application": app,
		"name":        pipelineName,
	}
	return api.Front50API.CreatePipeline(p)
}

// GetPipelineID gets a pipeline's ID.
func (api *DefaultPipelineAPI) GetPipelineID(app, pipelineName string) (string, error) {
	ids, err := api.PipelineIDs(app)
	if err != nil {
		return "", err
	}
	id, exists := ids[pipelineName]
	if !exists {
		if err := api.createEmptyPipeline(app, pipelineName); err != nil {
			return "", err
		}
		return api.GetPipelineID(app, pipelineName)
	}

	return id, nil
}

// PipelineIDs returns the pipeline IDs keyed by name for an application.
func (api *DefaultPipelineAPI) PipelineIDs(app string) (map[string]string, error) {
	ids := map[string]string{}
	log.Info("Looking up existing pipelines")
	resp, err := api.Front50API.GetPipelinesForApplication(app)
	if err != nil {
		return nil, err
	}
	type pipeline struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var pipelines []pipeline
	util.ReadJSON(bytes.NewReader(resp), &pipelines)
	for _, p := range pipelines {
		ids[p.Name] = p.ID
	}
	return ids, nil
}

func (api *DefaultPipelineAPI) updatePipeline(p Pipeline) error {
	_, err := json.Marshal(p)
	if err != nil {
		return err
	}

	if id, exists := p["id"]; exists {
		err = api.Front50API.UpdatePipeline(id.(string), p)
	} else {
		err = api.Front50API.CreatePipeline(p)
	}

	if err != nil {
		return err
	}

	log.Info("Successfully posted pipeline to Spinnaker.")
	return nil
}
