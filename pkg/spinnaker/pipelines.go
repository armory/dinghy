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
func UpdatePipelines(p []Pipeline) (err error) {
	if len(p) == 0 {
		return
	}
	app := p[0].Application()
	if !applicationExists(app) {
		NewApplication("unknown@unknown.com", app)
	}
	ids, _ := pipelineIDs(app)
	log.Info("Found pipelines for ", app, ": ", ids)
	for _, pipeline := range p {
		// Add ids to existing pipelines
		if id, exists := ids[pipeline.Name()]; exists {
			log.Debug("Added id ", id, " to pipeline ", pipeline.Name())
			pipeline["id"] = id
		}
		log.Info("Updating pipeline: " + pipeline.Name())
		if settings.S.AutoLockPipelines == "true" {
			log.Debug("Locking pipeline ", pipeline.Name())
			pipeline.Lock()
		}
		err := updatePipeline(pipeline)
		if err != nil {
			log.Error("Could not post pipeline to Spinnaker ", err)
		}
	}
	return
}

func updatePipeline(p Pipeline) error {
	b, err := json.Marshal(p)
	if err != nil {
		log.Error("Could not marshal pipeline ", err)
		return err
	}
	log.Info("Potsing pipeline to Spinnaker: ", string(b))
	url := fmt.Sprintf(`%s/pipelines`, settings.S.SpinnakerAPIURL)
	_, err = postWithRetry(url, b)
	if err != nil {
		return err
	}
	log.Info("Successfully posted pipeline to Spinnaker.")
	return nil
}

// PipelineIDs returns the pipeline IDs keyed by name for an application.
func pipelineIDs(app string) (map[string]string, error) {
	ids := map[string]string{}
	url := fmt.Sprintf("%s/applications/%s/pipelineConfigs", settings.S.SpinnakerAPIURL, app)
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
