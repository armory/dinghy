package plank

import (
	"errors"
	"fmt"
	"time"
)

type DataSourcesType struct {
	Enabled  []string `json:"enabled" mapstructure:"enabled"`
	Disabled []string `json:"disabled" mapstructure:"disabled"`
}

// Application as returned from the Spinnaker API.
type Application struct {
	Name        string          `json:"name" mapstructure:"name"`
	Email       string          `json:"email" mapstructure:"email"`
	Description string          `json:"description,omitempty" mapstructure:"description"`
	User        string          `json:"user,omitempty" mapstructure:"user"`
	DataSources DataSourcesType `json:"dataSources,omitempty" mapstructure:"dataSources"`
}

// GetApplication returns the Application data struct for the
// given application name.
func (c *Client) GetApplication(name string) (*Application, error) {
	var app Application
	if err := c.Get(c.URLs["front50"]+"/v2/applications/"+name, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

// GetApplications returns all applications (you can see, at least)
func (c *Client) GetApplications() (*[]Application, error) {
	var apps []Application
	if err := c.Get(c.URLs["front50"]+"/v2/applications", &apps); err != nil {
		return nil, err
	}
	return &apps, nil
}

// CreateApplication does what it says.
func (c *Client) CreateApplication(a *Application) error {
	ref, err := c.CreateTask(a.Name, fmt.Sprintf("Create Application: %s", a.Name), a)
	if err != nil {
		return fmt.Errorf("could not create application - %v", err)
	}
	task, err := c.PollTaskStatus(ref.Ref)
	if err != nil || task.Status == "TERMINAL" {
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = "received status TERMINAL"
		}
		return errors.New(fmt.Sprintf("failed to create application: %s", errMsg))
	}

	// This really shouldn't have to be here, but after the task to create an
	// app is marked complete sometimes the object still doesn't exist. So
	// after doing the create, and getting back a completion, we still need
	// to poll till we find the app in order to make sure future operations will
	// succeed.
	err = c.pollAppConfig(a.Name)
	return err
}

// pollAppConfig isn't exposed because not sure it's worth exposing.  Just
// call GetApplication() if you're expecting it to be there.
func (c *Client) pollAppConfig(appName string) error {
	timer := time.NewTimer(c.retryIncrement)
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for range t.C {
		_, err := c.GetApplication(appName)
		if err == nil {
			return nil
		}
		select {
		case <-timer.C:
			return fmt.Errorf("timed out waiting for app to appear - %v", err)
		default:
		}
	}
	return errors.New("exited poll loop before completion")
}
