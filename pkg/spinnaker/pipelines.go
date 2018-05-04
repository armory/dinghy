package spinnaker

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/settings"
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
func UpdatePipelines(app string, p []Pipeline, delStale bool) (err error) {
	if !applicationExists(app) {
		NewApplication("unknown@unknown.com", app)
	}
	ids, _ := PipelineIDs(app)
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
		if settings.S.AutoLockPipelines == "true" {
			log.Debug("Locking pipeline ", pipeline.Name())
			pipeline.Lock()
		}
		err := updatePipeline(pipeline)
		if err != nil {
			log.Error("Could not post pipeline to Spinnaker ", err)
			return err
		}
	}
	if delStale {
		// clear existing pipelines that weren't updated
		for id, updated := range checklist {
			if !updated {
				DeletePipeline(app, idToName[id])
			}
		}
	}
	return err
}

func createEmptyPipeline(app, pipelineName string) error {
	p := Pipeline{
		"application": app,
		"name":        pipelineName,
	}

	b, err := json.Marshal(p)
	if err != nil {
		return err
	}

	log.Info("Posting pipeline to Spinnaker: ", string(b))
	url := fmt.Sprintf(`%s/pipelines`, settings.S.Front50.BaseURL)
	_, err = postWithRetry(url, b)

	return err
}

// GetPipelineID gets a pipeline's ID.
func GetPipelineID(app, pipelineName string) (string, error) {
	ids, err := PipelineIDs(app)
	if err != nil {
		return "", err
	}
	id, exists := ids[pipelineName]
	if !exists {
		if err := createEmptyPipeline(app, pipelineName); err != nil {
			return "", err
		}
		return GetPipelineID(app, pipelineName)
	}

	return id, nil
}

func updatePipeline(p Pipeline) error {
	b, err := json.Marshal(p)
	if err != nil {
		log.Error("Could not marshal pipeline ", err)
		return err
	}

	if id, exists := p["id"]; exists {
		log.Info("Updating existing pipeline: ", string(b))
		url := fmt.Sprintf(`%s/pipelines/%s`, settings.S.Front50.BaseURL, id)
		_, err = putWithRetry(url, b)
	} else {
		log.Info("Posting pipeline to Spinnaker: ", string(b))
		url := fmt.Sprintf(`%s/pipelines`, settings.S.Front50.BaseURL)
		_, err = postWithRetry(url, b)
	}

	if err != nil {
		return err
	}
	log.Info("Successfully posted pipeline to Spinnaker.")
	return nil
}

// DeletePipeline deletes a pipeline
func DeletePipeline(app string, pipelineName string) error {
	url := fmt.Sprintf("%s/pipelines/%s/%s", settings.S.Front50.BaseURL, app, pipelineName)
	if _, err := deleteWithRetry(url); err != nil {
		return err
	}
	return nil
}

// PipelineIDs returns the pipeline IDs keyed by name for an application.
func PipelineIDs(app string) (map[string]string, error) {
	ids := map[string]string{}
	url := fmt.Sprintf("%s/pipelines/%s", settings.S.Front50.BaseURL, app)
	log.Info("Looking up existing pipelines")
	resp, err := getWithRetry(url)
	if err != nil {
		return ids, err
	}
	type pipeline struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var pipelines []pipeline
	util.ReadJSON(resp.Body, &pipelines)
	for _, p := range pipelines {
		ids[p.Name] = p.ID
	}
	return ids, nil
}
