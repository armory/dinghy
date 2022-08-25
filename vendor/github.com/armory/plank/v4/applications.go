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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"time"
)

// DataSourcesType creates this block:
//   "dataSources": {
//     "disabled": [],
//     "enabled": ["canaryConfigs"]
//   }
type DataSourcesType struct {
	Enabled  []string `json:"enabled" mapstructure:"enabled" yaml:"enabled" hcl:"enabled"`
	Disabled []string `json:"disabled" mapstructure:"disabled" yaml:"disabled" hcl:"disabled"`
}

// PermissionsType creates this block:
//   "permissions": {
//     "READ": ["armory-io", "core"],
//     "WRITE": ["armory-io", "core"]
//   }
type PermissionsType struct {
	Read    []string `json:"READ" mapstructure:"READ" yaml:"READ" hcl:"READ"`
	Write   []string `json:"WRITE" mapstructure:"WRITE" yaml:"WRITE" hcl:"WRITE"`
	Execute []string `json:"EXECUTE" mapstructure:"EXECUTE" yaml:"EXECUTE" hcl:"EXECUTE"`
}

// Front50 needs this struct to update permissions
type Front50Permissions struct {
	Name        string           `json:"name" mapstructure:"name" yaml:"name" hcl:"name"`
	Permissions *PermissionsType `json:"permissions,omitempty" mapstructure:"permissions" yaml:"permissions,omitempty" hcl:"permissions,omitempty"`
}

type NotificationsType map[string]interface{}

// Application as returned from the Spinnaker API.
type Application struct {
	Name           string                 `json:"name" mapstructure:"name" yaml:"name" hcl:"name"`
	Email          string                 `json:"email" mapstructure:"email" yaml:"email" hcl:"email"`
	Description    string                 `json:"description,omitempty" mapstructure:"description" yaml:"description,omitempty" hcl:"description,omitempty"`
	User           string                 `json:"user,omitempty" mapstructure:"user" yaml:"user,omitempty" hcl:"user,omitempty"`
	DataSources    *DataSourcesType       `json:"dataSources,omitempty" mapstructure:"dataSources" yaml:"datasources,omitempty" hcl:"datasources,omitempty"`
	Permissions    *PermissionsType       `json:"permissions,omitempty" mapstructure:"permissions" yaml:"permissions,omitempty" hcl:"permissions,omitempty"`
	Notifications  NotificationsType      `json:"notifications,omitempty" mapstructure:"notifications" yaml:"notifications,omitempty" hcl:"notifications,omitempty"`
	AppMetadata    map[string]interface{} `json:"appmetadata,omitempty" mapstructure:"appmetadata" yaml:"appmetadata,omitempty" hcl:"appmetadata,omitempty"`
	RepoSlug       string                 `json:"repoSlug,omitempty" mapstructure:"repoSlug" yaml:"repoSlug,omitempty" hcl:"repoSlug,omitempty"`
	RepoType       string                 `json:"repoType,omitempty" mapstructure:"repoType" yaml:"repoType,omitempty" hcl:"repoType,omitempty"`
	RepoProjectKey string                 `json:"repoProjectKey,omitempty" mapstructure:"repoProjectKey" yaml:"repoProjectKey,omitempty" hcl:"repoProjectKey,omitempty"`
}

// All in Metadata will be in root
func (a Application) MarshalJSON() ([]byte, error) {
	amap := make(map[string]interface{})
	mapstructure.Decode(a, &amap)
	delete(amap, "appmetadata")
	for key, val := range a.AppMetadata {
		amap[key] = val
	}
	return json.Marshal(amap)
}

// GetApplication returns the Application data struct for the
// given application name.
func (c *Client) GetApplication(name, traceparent string) (*Application, error) {
	var app Application
	if c.UseGate {
		if err := c.Get(c.URLs["gate"]+"/plank/v2/applications/"+name, traceparent, &app); err != nil {
			return nil, err
		}
	} else {
		if err := c.Get(c.URLs["front50"]+"/v2/applications/"+name, traceparent, &app); err != nil {
			return nil, err
		}
	}
	return &app, nil
}

// GetApplications returns all applications (you can see, at least)
func (c *Client) GetApplications(traceparent string) (*[]Application, error) {
	var apps []Application
	if c.UseGate {
		if err := c.Get(c.URLs["gate"]+"/plank/v2/applications", traceparent, &apps); err != nil {
			return nil, err
		}
	} else {
		if err := c.Get(c.URLs["front50"]+"/v2/applications", traceparent, &apps); err != nil {
			return nil, err
		}
	}
	return &apps, nil
}

// DeleteApplication deletes an application from the configured front50 store.
func (c *Client) DeleteApplication(name, traceparent string) error {
	return c.Delete(fmt.Sprintf("%s/v2/applications/%s", c.URLs["front50"], name), traceparent)
}

// UpdateApplication updates an application in the configured front50 store.
func (c *Client) UpdateApplication(app Application, traceparent string) error {
	var unused interface{}
	if c.UseGate {
		if err := c.PatchWithRetry(fmt.Sprintf("%s/plank/v2/applications/%s", c.URLs["gate"], app.Name), traceparent, ApplicationJson, app, &unused); err != nil {
			return fmt.Errorf("could not update application %q: %w", app.Name, err)
		}
	} else {
		if err := c.PatchWithRetry(fmt.Sprintf("%s/v2/applications/%s", c.URLs["front50"], app.Name), traceparent, ApplicationJson, app, &unused); err != nil {
			return fmt.Errorf("could not update application %q: %w", app.Name, err)
		}
	}

	if err := c.UpdatePermissions(app.Name, traceparent, app.Permissions); err != nil {
		// UpdatePermissions will print in the log if something failed
		return err
	}

	return nil
}

// UpdatePermissions updates an application permissions in the configured front50 store.
func (c *Client) UpdatePermissions(appName, traceparent string, permissions *PermissionsType) error {
	var unused interface{}
	var permissionsfront50 = Front50Permissions{
		Name:        appName,
		Permissions: permissions,
	}
	if err := c.PutWithRetry(fmt.Sprintf("%s/permissions/applications/%s", c.URLs["front50"], appName), traceparent, ApplicationJson, permissionsfront50, &unused); err != nil {
		return fmt.Errorf("could not update application permissions %q: %w", appName, err)
	}
	return nil
}

type createApplicationTask struct {
	Application Application `json:"application" mapstructure:"application" yaml:"application" hcl:"application"`
	Type        string      `json:"type" mapstructure:"type" yaml:"type" hcl:"type"`
}

// CreateApplication does what it says.
func (c *Client) CreateApplication(a *Application, traceparent string) error {
	payload := createApplicationTask{Application: *a, Type: "createApplication"}
	ref, err := c.CreateTask(a.Name, fmt.Sprintf("Create Application: %s", a.Name), traceparent, payload)
	if err != nil {
		return fmt.Errorf("could not create application - %v", err)
	}
	task, err := c.PollTaskStatus(ref.Ref, traceparent)
	if err != nil || task.Status == "TERMINAL" {
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = "received status TERMINAL"
		}
		return errors.New(fmt.Sprintf("failed to create application: %s", errMsg))
	}

	// Not worried if ResyncFiat fails -- if ArmoryEndpoints not enabled, this
	// is a no-op, if it fails, the polling later might still succeed, or we'll
	// get an error about not being able to retrieve the Application.
	c.ResyncFiat("")

	// This really shouldn't have to be here, but after the task to create an
	// app is marked complete sometimes the object still doesn't exist. So
	// after doing the create, and getting back a completion, we still need
	// to poll till we find the app in order to make sure future operations will
	// succeed.
	err = c.pollAppConfig(a.Name, traceparent)
	return err
}

// pollAppConfig isn't exposed because not sure it's worth exposing.  Just
// call GetApplication() if you're expecting it to be there.
func (c *Client) pollAppConfig(appName, traceparent string) error {
	timer := time.NewTimer(4 * time.Minute)
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for range t.C {
		_, err := c.GetApplication(appName, traceparent)
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
