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
	"fmt"
)

// Pipeline is the structure that comes back from Spinnaker
// representing a pipeline definition (different than an execution)
type Pipeline struct {
	ID                   string                   `json:"id,omitempty"`
	Type                 string                   `json:"type,omitempty"`
	Name                 string                   `json:"name"`
	Application          string                   `json:"application"`
	Description          string                   `json:"description,omitempty"`
	ExecutionEngine      string                   `json:"executionEngine,omitempty"`
	Parallel             bool                     `json:"parallel"`
	LimitConcurrent      bool                     `json:"limitConcurrent"`
	KeepWaitingPipelines bool                     `json:"keepWaitingPipelines"`
	Stages               []map[string]interface{} `json:"stages,omitempty"`
	Triggers             []map[string]interface{} `json:"triggers,omitempty"`
	Parameters           []map[string]interface{} `json:"parameterConfig,omitempty"`
	Notifications        []map[string]interface{} `json:"notifications,omitempty"`
	LastModifiedBy       string                   `json:"lastModifiedBy"`
	Config               interface{}              `json:"config,omitempty"`
	UpdateTs             string                   `json:"updateTs"`
	Locked               PipelineLockType         `json:"locked,omitempty"`
}

type PipelineLockType struct {
	UI            bool `json:"ui"`
	AllowUnlockUI bool `json:"allowUnlockUi"`
}

func (p *Pipeline) Lock() *Pipeline {
	p.Locked.UI = true
	p.Locked.AllowUnlockUI = true
	return p
}

// utility to return the base URL for all pipelines API calls
func (c *Client) pipelinesURL() string {
	return c.URLs["front50"] + "/pipelines"
}

// Get returns an array of all the Spinnaker pipelines
// configured for app
func (c *Client) GetPipelines(app string) ([]Pipeline, error) {
	var pipelines []Pipeline
	if err := c.GetWithRetry(c.pipelinesURL() + "/" + app, &pipelines); err != nil {
		return nil, fmt.Errorf("could not get pipelines for %s - %v", app, err)
	}
	return pipelines, nil
}

// UpsertPipeline creates/updates a pipeline defined in the struct argument.
func (c *Client) UpsertPipeline(p Pipeline, id string) error {
	var unused interface{}
	if id == "" {
		if err := c.PostWithRetry(c.pipelinesURL(), ApplicationJson, p, &unused); err != nil {
			return fmt.Errorf("could not create pipeline - %v", err)
		}
	} else {
		if err := c.PutWithRetry(fmt.Sprintf("%s/%s", c.pipelinesURL(), id), ApplicationJson, p, &unused); err != nil {
			return fmt.Errorf("could not update pipeline - %v", err)
		}
	}
	return nil
}

// DeletePipeline does what it says.
func (c *Client) DeletePipeline(p Pipeline) error {
	return c.DeleteWithRetry(
		fmt.Sprintf("%s/%s/%s", c.pipelinesURL(), p.Application, p.Name))
}

func (c *Client) DeletePipelineByName(app, pipeline string) error {
	return c.DeleteWithRetry(
		fmt.Sprintf("%s/%s/%s", c.pipelinesURL(), app, pipeline))
}

type pipelineExecution struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	DryRun  bool   `json:"dryRun"`
	User    string `json:"user"`
}

type PipelineRef struct {
	// Ref is the path the the execution. Use it to get status updates.
	Ref string `json:"ref"`
}

// Execute a pipeline by application and pipeline.
func (c *Client) Execute(application, pipeline string) (*PipelineRef, error) {
	e := pipelineExecution{
		Enabled: true,
		Type:    "manual",
		DryRun:  false,
		User:    "anonymous",
	}
	var ref PipelineRef
	if err := c.PostWithRetry(
		fmt.Sprintf("%s/%s/%s", c.pipelinesURL(), application, pipeline),
		ApplicationJson, e, &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}
