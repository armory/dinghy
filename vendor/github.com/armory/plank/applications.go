/*
 * Copyright 2019 Armory, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package plank

import (
	"errors"
	"fmt"
	"time"
)

type DataSourcesType struct {
	Enabled  []string `json:"enabled" mapstructure:"enabled" yaml:"enabled" hcl:"enabled"`
	Disabled []string `json:"disabled" mapstructure:"disabled" yaml:"disabled" hcl:"disabled"`
}

// Application as returned from the Spinnaker API.
type Application struct {
	Name        string          `json:"name" mapstructure:"name" yaml:"name" hcl:"name"`
	Email       string          `json:"email" mapstructure:"email" yaml:"email" hcl:"email"`
	Description string          `json:"description,omitempty" mapstructure:"description" yaml:"description,omitempty" hcl:"description,omitempty"`
	User        string          `json:"user,omitempty" mapstructure:"user" yaml:"user,omitempty" hcl:"user,omitempty"`
	DataSources DataSourcesType `json:"dataSources,omitempty" mapstructure:"dataSources" yaml:"datasources,omitempty" hcl:"datasources,omitempty"`
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

type createApplicationTask struct {
	Application Application `json:"application" mapstructure:"application" yaml:"application" hcl:"application"`
	Type        string      `json:"type" mapstructure:"type" yaml:"type" hcl:"type"`
}

// CreateApplication does what it says.
func (c *Client) CreateApplication(a *Application) error {
	payload := createApplicationTask{Application: *a, Type: "createApplication"}
	ref, err := c.CreateTask(a.Name, fmt.Sprintf("Create Application: %s", a.Name), payload)
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
