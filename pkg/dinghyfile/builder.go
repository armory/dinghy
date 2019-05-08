/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package dinghyfile

import (
	"errors"

	"github.com/armory-io/dinghy/pkg/events"
	log "github.com/sirupsen/logrus"

	"github.com/armory/plank"
)

// PipelineBuilder is responsible for downloading dinghyfiles/modules, compiling them, and sending them to Spinnaker
type PipelineBuilder struct {
	Downloader           Downloader
	Depman               DependencyManager
	TemplateRepo         string
	TemplateOrg          string
	DinghyfileName       string
	Client               *plank.Client
	DeleteStalePipelines bool
	AutolockPipelines    string
	EventClient          events.EventClient
}

// DependencyManager is an interface for assigning dependencies and looking up root nodes
type DependencyManager interface {
	SetDeps(parent string, deps []string)
	GetRoots(child string) []string
}

// Downloader is an interface that fetches files from a source
type Downloader interface {
	Download(org, repo, file string) (string, error)
	EncodeURL(org, repo, file string) string
	DecodeURL(url string) (string, string, string)
}

// Dinghyfile is the format of the pipeline template JSON
type Dinghyfile struct {
	// Application name can be specified either in top-level "application" or as a key in "spec"
	// We don't want arbitrary application properties in the top-level Dinghyfile so we put them in .spec
	Application          string                 `json:"application"`
	ApplicationSpec      plank.Application      `json:"spec"`
	DeleteStalePipelines bool                   `json:"deleteStalePipelines"`
	Globals              map[string]interface{} `json:"globals"`
	Pipelines            []plank.Pipeline       `json:"pipelines"`
}

func NewDinghyfile() Dinghyfile {
	return Dinghyfile{
		// initialize the application spec so that the default
		// enabled/disabled are initilzed slices
		// https://danott.co/posts/json-marshalling-empty-slices-to-empty-arrays-in-go.html
		ApplicationSpec: plank.Application{
			DataSources: plank.DataSourcesType{
				Enabled:  []string{},
				Disabled: []string{},
			},
		},
	}
}

var (
	// ErrMalformedJSON is more specific than just returning 422.
	ErrMalformedJSON = errors.New("malformed json")
	DefaultEmail     = "unknown@unknown.com"
)

func UpdateDinghyfile(dinghyfile []byte) (Dinghyfile, error) {
	d := NewDinghyfile()
	if err := Unmarshal(dinghyfile, &d); err != nil {
		return d, ErrMalformedJSON
	}
	log.Info("Unmarshalled: ", d)

	// If "spec" is not provided, these will be initialized to ""; need to pull them in.
	if d.ApplicationSpec.Name == "" {
		d.ApplicationSpec.Name = d.Application
	}
	if d.ApplicationSpec.Email == "" {
		d.ApplicationSpec.Email = DefaultEmail
	}

	return d, nil
}

// ProcessDinghyfile downloads a dinghyfile and uses it to update Spinnaker's pipelines.
func (b *PipelineBuilder) ProcessDinghyfile(org, repo, path string) error {
	// Render the dinghyfile and decode it into a Dinghyfile object
	buf, err := b.Render(org, repo, path, nil)
	if err != nil {
		log.Errorf("Failed to render dinghyfile %s: %s", path, err.Error())
		return err
	}
	log.Debug("Rendered: ", buf.String())
	d, err := UpdateDinghyfile(buf.Bytes())
	if err != nil {
		log.Errorf("Failed to update dinghyfile %s: %s", path, err.Error())
		return err
	}
	log.Debug("Updated: ", buf.String())

	if err := b.updatePipelines(&d.ApplicationSpec, d.Pipelines, d.DeleteStalePipelines, b.AutolockPipelines); err != nil {
		log.Errorf("Failed to update Pipelines for %s: %s", path, err.Error())
		return err
	}

	return nil
}

// RebuildModuleRoots rebuilds all dinghyfiles which are roots of the specified file
func (b *PipelineBuilder) RebuildModuleRoots(org, repo, path string) error {
	errEncountered := false
	failedUpdates := []string{}
	url := b.Downloader.EncodeURL(org, repo, path)
	log.Info("Processing module: " + url)

	// TODO: could handle logging and errors for file processing more elegantly rather
	// than making two passes.
	// Process all dinghyfiles that depend on this module
	for _, url := range b.Depman.GetRoots(url) {
		// TODO: we don't need to decode here because these values come in as parameters
		org, repo, path := b.Downloader.DecodeURL(url)
		if err := b.ProcessDinghyfile(org, repo, path); err != nil {
			errEncountered = true
			failedUpdates = append(failedUpdates, url)
		}
	}

	if errEncountered {
		log.Error("The following dinghyfiles weren't updated successfully:")
		for d := range failedUpdates {
			log.Error(d)
		}
		return errors.New("Not all upstream dinghyfiles were updated successfully")
	}
	return nil
}

func (b *PipelineBuilder) updatePipelines(app *plank.Application, pipelines []plank.Pipeline, deleteStale bool, autoLock string) error {
	_, err := b.Client.GetApplication(app.Name)
	if err != nil {
		// Likely just not there...
		log.Infof("Creating application '%s'...", app.Name)
		if err := b.Client.CreateApplication(app); err != nil {
			log.Errorf("Failed to create application (%s)", err.Error())
			return err
		}
	}

	ids, _ := b.PipelineIDs(app.Name)
	ignoreList := make(map[string]bool)
	idToName := make(map[string]string)
	for name, id := range ids {
		ignoreList[id] = false
		idToName[id] = name
	}
	log.Info("Found pipelines for ", app, ": ", ids)
	for _, p := range pipelines {
		// Add ids to existing pipelines
		if id, exists := ids[p.Name]; exists {
			log.Debug("Added id ", id, " to pipeline ", p.Name)
			ignoreList[p.Name] = true
			p.ID = id //note: we're working with a copy.  once this loop exits all changes go out of scope!
			log.Info("Updating pipeline: " + p.Name)
		} else {
			log.Debug("Adding ", p.Name, " to ignored stale pipelines")
			ignoreList[p.Name] = true
			log.Info("Creating pipeline: " + p.Name)
		}
		if autoLock == "true" {
			log.Debug("Locking pipeline ", p.Name)
			p.Lock()
		}
		if err := b.Client.UpsertPipeline(p, p.ID); err != nil {
			log.Errorf("Upsert failed: %s", err.Error())
			return err
		}
		log.Info("Upsert succeeded.")
	}
	if deleteStale {
		// clear existing pipelines that weren't updated
		log.Debug("Pipelines we should ignore because they were just created: ", ignoreList)
		for _, p := range pipelines {
			if !ignoreList[p.Name] {
				log.Infof("Deleting stale pipeline %s", p.Name)
				if err := b.Client.DeletePipeline(p); err != nil {
					// Not worrying about handling errors here because it just means it
					// didn't get deleted *this time*.
					log.Warnf("Could not delete Pipeline %s (Application %s)", p.Name, p.Application)
				}
			}
		}
	}
	return err
}

func (b *PipelineBuilder) PipelineIDs(app string) (map[string]string, error) {
	ids := map[string]string{}
	log.Info("Looking up existing pipelines")
	pipelines, err := b.Client.GetPipelines(app)
	if err != nil {
		return ids, err
	}
	for _, p := range pipelines {
		ids[p.Name] = p.ID
	}
	return ids, nil
}
