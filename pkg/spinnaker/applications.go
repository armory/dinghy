package spinnaker

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/util"
)

// Application represents an applications configuration.
type Application struct {
	Name string `json:"name"`
}

// CreateApplicationJob is the element to include in the Job array of a Task
// request when creating a new app
type createApplicationJob struct {
	Application applicationTaskAttributes `json:"application"`
	Type        string                    `json:"type"`
	User        string                    `json:"user,omitempty"`
}

type applicationTaskAttributes struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name"`
}

// Applications returns a list of applications
func Applications() []string {
	ret := make([]string, 0)
	url := fmt.Sprintf("%s/applications", settings.SpinnakerAPIURL)
	resp, err := getWithRetry(url)
	if err != nil {
		log.Info("Failed to get application.")
		return []string{}
	}
	var apps []Application
	util.ReadJSON(resp.Body, &apps)
	for _, app := range apps {
		ret = append(ret, app.Name)
	}
	resp.Body.Close()
	return ret
}

func applicationExists(a string) bool {
	apps := Applications()
	for _, app := range apps {
		if strings.ToLower(app) == strings.ToLower(a) {
			return true
		}
	}
	return false
}

// NewApplication creates an app on Spinnaker, but that's an async
// request made with the tasks interface. So we submit the task, and poll for
// the task completion.
func NewApplication(email string, name string) (err error) {
	createAppJob := createApplicationJob{
		Application: applicationTaskAttributes{
			Email: email,
			Name:  name,
		},
		Type: "createApplication",
	}
	createApp := Task{
		Application: name,
		Description: "Create Application: " + name,
		Job:         []interface{}{createAppJob},
	}

	log.Info("Creating application " + name)
	ref, err := submitTask(createApp)
	if err != nil {
		return err
	}

	log.Info("Polling for app create complete at " + ref.Ref)
	resp, err := pollTaskStatus(ref.Ref, 4*time.Minute)
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
	err = pollAppConfig(name, 7*time.Minute)
	if err != nil {
		return fmt.Errorf("Couldn't find app after creation: %v", err)
	}
	return nil
}

func pollAppConfig(app string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for range t.C {
		if applicationExists(app) {
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
