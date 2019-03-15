package spinnaker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/util"
)

// Application represents an applications configuration.
type Application struct {
	Name string `json:"name"`
}

// createApplicationJob is the element to include in the Job array of a Task
// request when creating a new app
type createApplicationJob struct {
	Application ApplicationSpec `json:"application"`
	Type        string          `json:"type"`
	User        string          `json:"user,omitempty"`
}

type DataSourcesSpec struct {
	Enabled  []string `json:"enabled"`
	Disabled []string `json:"disabled"`
}

// Must include `name` and `email`
// DataSources must be a pointer otherwise we end up sending nil/empty dicts, which break the UI
type ApplicationSpec struct {
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Description string          `json:"description,omitempty"`
	User        string          `json:"user,omitempty"`
	DataSources DataSourcesSpec `json:"dataSources,omitempty"`
}

type Front50API interface {
	Applications() []string
	NewApplication(spec ApplicationSpec) error
	ApplicationExists(appName string) bool
	CreatePipeline(p Pipeline) error
	DeletePipeline(appName, pipelineName string) error
	UpdatePipeline(id string, pipeline Pipeline) error
	GetPipelinesForApplication(appName string) ([]byte, error)
}

type DefaultFront50API struct {
	BaseURL   string
	OrcaAPI   OrcaAPI
	APIClient APIClient
}

// TODO: this function has errors, should return them and let the caller handle
// Applications returns a list of applications
func (df50a *DefaultFront50API) Applications() []string {
	ret := make([]string, 0)
	url := fmt.Sprintf("%s/v2/applications", df50a.BaseURL)
	resp, err := df50a.APIClient.GetWithRetry(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		// TODO: return this error
		log.Info("Failed to get application.")
		return []string{}
	}
	var apps []Application
	util.ReadJSON(resp.Body, &apps)
	for _, app := range apps {
		ret = append(ret, app.Name)
	}
	return ret
}

// NewApplication creates an app on Spinnaker, but that's an async
// request made with the tasks interface. So we submit the task, and poll for
// the task completion.
func (df50a *DefaultFront50API) NewApplication(spec ApplicationSpec) (err error) {
	name := spec.Name

	createAppJob := createApplicationJob{
		Application: spec,
		Type:        "createApplication",
		User:        "Dinghy", // TODO update this with an actual user
	}
	createApp := Task{
		Application: name,
		Description: "Create Application from Dinghy: " + name,
		Job:         []interface{}{createAppJob},
	}

	log.Info("Creating application " + name)
	ref, err := df50a.OrcaAPI.SubmitTask(createApp)
	if err != nil {
		return err
	}

	log.Info("Polling for app create complete at " + ref.Ref)
	resp, err := df50a.OrcaAPI.PollTaskStatus(ref.Ref, 4*time.Minute)
	if err != nil {
		return fmt.Errorf("Error polling app create %s: %v", name, err)
	}

	if resp.Status == "TERMINAL" {
		log.WithField("status", resp.Status).Error("Task failed")
		if retrofitErr := resp.ExtractRetrofitError(); retrofitErr != nil {
			return errors.New("Retrofit error: " + retrofitErr.ResponseBody)
		}
		return errors.New("Unknown Retrofit error: " + fmt.Sprintf("%#v", resp))
	}
	log.WithField("status", resp.Status).Info("Task completed")

	// This really shouldn't have to be here, but after the task to create an
	// app is marked complete sometimes the object still doesn't exist. So
	// after doing the create, and getting back a completion, we still need
	// to poll till we find the app in order to make sure future operations will
	// succeed.
	err = df50a.pollAppConfig(name, 7*time.Minute)
	if err != nil {
		return fmt.Errorf("Couldn't find app after creation: %v", err)
	}
	return nil
}

func (df50a *DefaultFront50API) ApplicationExists(a string) bool {
	apps := df50a.Applications()
	for _, app := range apps {
		if strings.ToLower(app) == strings.ToLower(a) {
			return true
		}
	}
	return false
}

func (df50a *DefaultFront50API) CreatePipeline(p Pipeline) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(`%s/pipelines`, df50a.BaseURL)
	resp, err := df50a.APIClient.PostWithRetry(url, b)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}

func (df50a *DefaultFront50API) DeletePipeline(appName, pipelineName string) error {
	url := fmt.Sprintf("%s/pipelines/%s/%s", df50a.BaseURL, appName, pipelineName)
	resp, err := df50a.APIClient.DeleteWithRetry(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	return nil
}

func (df50a *DefaultFront50API) UpdatePipeline(id string, pipeline Pipeline) error {
	b, err := json.Marshal(pipeline)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(`%s/pipelines`, df50a.BaseURL)
	resp, err := df50a.APIClient.PostWithRetry(url, b)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}

func (df50a *DefaultFront50API) GetPipelinesForApplication(appName string) ([]byte, error) {
	url := fmt.Sprintf("%s/pipelines/%s", df50a.BaseURL, appName)
	resp, err := df50a.APIClient.GetWithRetry(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func (df50a *DefaultFront50API) pollAppConfig(app string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for range t.C {
		if df50a.ApplicationExists(app) {
			return nil
		}

		select {
		case <-timer.C:
			return errors.New("timed out waiting for app to appear")
		default:
			log.Debug("Polling...")
		}
	}
	return errors.New("exited poll loop before completion")
}
