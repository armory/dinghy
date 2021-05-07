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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/armory/dinghy/pkg/dinghyfile/pipebuilder"
	"github.com/armory/dinghy/pkg/log"
	"path/filepath"
	"regexp"
	"time"

	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/notifiers"
	"github.com/armory/dinghy/pkg/util"
	"github.com/armory/plank/v4"
)

type VarMap map[string]interface{}

type Parser interface {
	SetBuilder(b *PipelineBuilder)
	Parse(org, repo, path, branch string, vars []VarMap) (*bytes.Buffer, error)
}

// PipelineBuilder is responsible for downloading dinghyfiles/modules, compiling them, and sending them to Spinnaker
type PipelineBuilder struct {
	Downloader                  Downloader
	Depman                      DependencyManager
	TemplateRepo                string
	TemplateOrg                 string
	DinghyfileName              string
	Client                      util.PlankClient
	DeleteStalePipelines        bool
	AutolockPipelines           string
	EventClient                 events.EventClient
	Parser                      Parser
	Logger                      log.DinghyLog
	Ums                         []Unmarshaller
	Notifiers                   []notifiers.Notifier
	PushRaw                     map[string]interface{}
	GlobalVariablesMap          map[string]interface{}
	RepositoryRawdataProcessing bool
	RebuildingModules           bool
	Action                      pipebuilder.BuilderAction
}

// DependencyManager is an interface for assigning dependencies and looking up root nodes
type DependencyManager interface {
	GetRawData(url string) (string, error)
	SetRawData(url string, rawData string) error
	SetDeps(parent string, deps []string)
	GetRoots(child string) []string
}

// Downloader is an interface that fetches files from a source
type Downloader interface {
	Download(org, repo, file, branch string) (string, error)
	EncodeURL(org, repo, file, branch string) string
	DecodeURL(url string) (string, string, string, string)
}

// Dinghyfile is the format of the pipeline template JSON
type Dinghyfile struct {
	// Application name can be specified either in top-level "application" or as a key in "spec"
	// We don't want arbitrary application properties in the top-level Dinghyfile so we put them in .spec
	Application          string                 `json:"application" yaml:"application" hcl:"application"`
	ApplicationSpec      plank.Application      `json:"spec" yaml:"spec" hcl:"spec"`
	DeleteStalePipelines bool                   `json:"deleteStalePipelines" yaml:"deleteStalePipelines" hcl:"deleteStalePipelines"`
	Globals              map[string]interface{} `json:"globals" yaml:"globals" hcl:"globals"`
	Pipelines            []plank.Pipeline       `json:"pipelines" yaml:"pipelines" hcl:"pipelines"`
}

func NewDinghyfile() Dinghyfile {
	return Dinghyfile{
		// initialize the application spec so that the default
		// enabled/disabled are initilzed slices
		// https://danott.co/posts/json-marshalling-empty-slices-to-empty-arrays-in-go.html
		ApplicationSpec: plank.Application{
			DataSources: &plank.DataSourcesType{
				Enabled:  []string{},
				Disabled: []string{},
			},
		},
	}
}

var (
	// ErrMalformedJSON is more specific than just returning 422.
	ErrMalformedJSON = errors.New("malformed json")
	DefaultEmail     = "unknown@example.org"
)

// UpdateDinghyfile doesn't actually update anything; it unmarshals the
// results bytestream into the Dinghyfile{} struct, and sets the name and
// email on the ApplicationSpec if it didn't get unmarshalled on its own.
func (b *PipelineBuilder) UpdateDinghyfile(dinghyfile []byte) (Dinghyfile, error) {
	d := NewDinghyfile()
	// try every parser, maybe we'll get lucky
	parseErrs := 0
	for _, ums := range b.Ums {
		if err := ums.Unmarshal(dinghyfile, &d); err != nil {
			b.Logger.Warnf("UpdateDinghyfile malformed syntax: %s", err.Error())
			parseErrs++
			continue
		}
	}
	event := &events.Event{
		Start:      time.Now().UTC().Unix(),
		End:        time.Now().UTC().Unix(),
		Org:        "",
		Repo:       "",
		Path:       "",
		Branch:     "",
		Dinghyfile: string(dinghyfile),
		Module:     false,
	}

	// we weren't lucky, all the parsers failed
	if parseErrs == len(b.Ums) {
		b.Logger.Errorf("update-dinghyfile-unmarshal-err: %s", string(dinghyfile))
		b.EventClient.SendEvent("update-dinghyfile-unmarshal-err", event)
		return d, ErrMalformedJSON
	}
	b.Logger.Infof("Unmarshalled: %v", d)

	// If "spec" is not provided, these will be initialized to ""; need to pull them in.
	if d.ApplicationSpec.Name == "" {
		d.ApplicationSpec.Name = d.Application
	}
	if d.ApplicationSpec.Email == "" {
		d.ApplicationSpec.Email = DefaultEmail
	}

	return d, nil
}

// We will validate using plank refs at this moment
func (b *PipelineBuilder) ValidatePipelines(d Dinghyfile, dinghyfile []byte) error {

	event := &events.Event{
		Start:      time.Now().UTC().Unix(),
		End:        time.Now().UTC().Unix(),
		Org:        "",
		Repo:       "",
		Path:       "",
		Branch:     "",
		Dinghyfile: string(dinghyfile),
		Module:     false,
	}

	var lastErr error
	var warning bool
	for _, pipeline := range d.Pipelines {
		validateResult := pipeline.ValidateRefIds()
		for _, stageWarning := range validateResult.Warnings {
			warning = true
			b.Logger.Warnf("There are some concerns validating stage refs for pipeline: %s", stageWarning)
		}
		for _, stageError := range validateResult.Errors {
			lastErr = stageError
			b.Logger.Errorf("Failed to validate stage refs for pipeline: %s", stageError.Error())
		}
	}
	if warning {
		b.EventClient.SendEvent("validate-pipelines-stagerefs-warn", event)
	}
	if lastErr != nil {
		b.Logger.Errorf("validate-pipelines-stagerefs-err: %s", string(dinghyfile))
		b.EventClient.SendEvent("validate-pipelines-stagerefs-err", event)
		return lastErr
	}
	return nil
}

// We will validate app notifications struc at this moment
func (b *PipelineBuilder) ValidateAppNotifications(d Dinghyfile, dinghyfile []byte) error {

	event := &events.Event{
		Start:      time.Now().UTC().Unix(),
		End:        time.Now().UTC().Unix(),
		Org:        "",
		Repo:       "",
		Path:       "",
		Branch:     "",
		Dinghyfile: string(dinghyfile),
		Module:     false,
	}

	err := d.ApplicationSpec.Notifications.ValidateAppNotification()
	if err != nil {
		b.Logger.Errorf("validate-app-notifications-err: %s", string(dinghyfile))
		b.EventClient.SendEvent("validate-app-notifications-err", event)
		return err
	}
	return nil
}

// DetermineParser currently only returns a DinghyfileParser; it could
// return other types of parsers in the future (for example, MPTv2)
// If we can't discern the types based on the path passed here, we may need
// to revisit this.  For now, this is just a stub that always returns the
// DinghyfileParser type.
func (b *PipelineBuilder) DetermineParser(path string) Parser {
	return NewDinghyfileParser(b)
}

// ProcessDinghyfile downloads a dinghyfile and uses it to update Spinnaker's pipelines.
func (b *PipelineBuilder) ProcessDinghyfile(org, repo, path, branch string) (string, error) {
	if b.Parser == nil {
		// Set the renderer based on evaluation of the path, if not already set
		b.Logger.Info("Calling DetermineParser")
		b.Parser = b.DetermineParser(path)
	}
	buf, err := b.Parser.Parse(org, repo, path, branch, nil)
	if err != nil {
		buf, errDownload := b.Downloader.Download(org, repo, path, branch)
		b.Logger.Errorf("Failed to parse dinghyfile %s: %s", path, err.Error())
		if errDownload == nil {
			b.NotifyFailure(org, repo, path, err, buf)
			return buf, err
		} else {
			b.NotifyFailure(org, repo, path, err, "")
			return "", err
		}
	}
	b.Logger.Infof("Compiled: %s", buf.String())
	d, err := b.UpdateDinghyfile(buf.Bytes())
	if err != nil {
		b.Logger.Errorf("Failed to update dinghyfile %s: %s", path, err.Error())
		b.NotifyFailure(org, repo, path, err, buf.String())
		return buf.String(), err
	}
	b.Logger.Infof("Updated: %s", buf.String())
	b.Logger.Infof("Dinghyfile struct: %v", d)

	err = b.ValidatePipelines(d, buf.Bytes())
	if err != nil {
		b.Logger.Errorf("Failed to validate pipelines %s", path)
		b.NotifyFailure(org, repo, path, err, buf.String())
		return buf.String(), err
	}
	b.Logger.Info("Validations for stage refs were successful")

	err = b.ValidateAppNotifications(d, buf.Bytes())
	if err != nil {
		b.Logger.Errorf("Failed to validate application notifications %s", d.ApplicationSpec.Notifications)
		b.NotifyFailure(org, repo, path, err, buf.String())
		return buf.String(), err
	}
	b.Logger.Info("Validations for app notifications were successful")

	if b.Action == pipebuilder.Validate {
		b.Logger.Info("Validation finished successfully")
	} else {
		if err := b.updatePipelines(&d.ApplicationSpec, d.Pipelines, d.DeleteStalePipelines, b.AutolockPipelines); err != nil {
			b.Logger.Errorf("Failed to update Pipelines for %s: %s", path, err.Error())
			b.NotifyFailure(org, repo, path, err, buf.String())
			return buf.String(), err
		}
	}

	b.NotifySuccess(org, repo, path, d.ApplicationSpec.Notifications)
	return buf.String(), nil
}

func unwrapFront50Error(err error) error {
	// Front50/OPA errors are in the form {"error": "BadRequest", "message": "foo"}
	// So let's attempt to destructure that value and report it downstream so
	// users have a nicer error message. If we can't parse the message for
	// whatever reason, then we'll just end up returning the previous error instead.
	var fr *plank.FailedResponse
	if errors.As(err, &fr) {
		var bre Front50BadRequestError
		if jsonErr := json.Unmarshal(fr.Response, &bre); jsonErr == nil {
			return &bre // Make sure this is the error that gets persisted now
		}
	}

	return err
}

// Front50BadRequestError represents a 4xx response from Front50
type Front50BadRequestError struct {
	Type   string `json:"error"`
	Reason string `json:"message"`
}

func (e *Front50BadRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Reason)
}

// RebuildModuleRoots rebuilds all dinghyfiles which are roots of the specified file
func (b *PipelineBuilder) RebuildModuleRoots(org, repo, path, branch string) error {
	b.RebuildingModules = true
	errEncountered := false
	failedUpdates := []string{}
	// if we are doing a update on template repo, we should test against the branch
	if b.Action == pipebuilder.Validate && b.TemplateRepo != repo {
		// Since we are checking for modules, those live in master
		branch = "master"
	}
	url := b.Downloader.EncodeURL(org, repo, path, branch)
	b.Logger.Info("Processing module: " + url)

	// TODO: could handle logging and errors for file processing more elegantly rather
	// than making two passes.
	// Process all dinghyfiles that depend on this module
	for _, url := range b.Depman.GetRoots(url) {
		org, repo, path, branch := b.Downloader.DecodeURL(url)
		if filepath.Base(path) == b.DinghyfileName {
			if b.RepositoryRawdataProcessing {
				rawData, errRaw := b.Depman.GetRawData(url)
				if errRaw == nil && rawData != "" {
					b.Logger.Infof("found rawdata for %v", url)
					// deserialze push data to a map.
					rawPushData := make(map[string]interface{})
					if err := json.Unmarshal([]byte(rawData), &rawPushData); err != nil {
						b.Logger.Errorf("unable to deserialize raw data to map while executing RebuildModuleRoots")
					} else {
						b.Logger.Infof("using latest rawdata from %v", url)
						b.PushRaw = rawPushData
					}
				}
			}

			if _, err := b.ProcessDinghyfile(org, repo, path, branch); err != nil {
				errEncountered = true
				failedUpdates = append(failedUpdates, url)
			}
		}

	}

	if errEncountered {
		var word string
		if b.Action == pipebuilder.Validate {
			word = "validated"
		} else {
			word = "updated"
		}
		b.Logger.Errorf("The following dinghyfiles weren't %v successfully:", word)
		for _, url := range failedUpdates {
			b.Logger.Error(url)
		}
		return fmt.Errorf("Not all upstream dinghyfiles were %v successfully", word)
	}
	return nil
}

// This is the bit that actually updates the pipeline(s) in Spinnaker
func (b *PipelineBuilder) updatePipelines(app *plank.Application, pipelines []plank.Pipeline, deleteStale bool, autoLock string) error {
	var newapp = false
	_, err := b.Client.GetApplication(app.Name, "")
	if err != nil {
		newapp = true
		failedResponse, ok := err.(*plank.FailedResponse)
		if !ok {
			b.Logger.Errorf("Failed to create application (%s)", err.Error())
			return err
		}
		if failedResponse.StatusCode == 404 {
			// Likely just not there...
			b.Logger.Infof("Creating application '%s'...", app.Name)
			if err = b.Client.CreateApplication(app, ""); err != nil {
				b.Logger.Errorf("Failed to create application (%s)", failedResponse.Error())
				return err
			}
		} else {
			b.Logger.Errorf("Failed to create application (%s)", failedResponse.Error())
			return err
		}
	} else {
		if val, found := b.GlobalVariablesMap["save_app_on_update"]; found && val == true {
			errUpdating := b.Client.UpdateApplication(*app, "")
			if errUpdating != nil {
				b.Logger.Errorf("Failed to update application (%s)", errUpdating.Error())
				return errUpdating
			}
		}
	}

	if val, found := b.GlobalVariablesMap["save_app_on_update"]; (found && val == true) || newapp {
		b.Logger.Infof("Updating notifications: %s", app.Notifications)
		errNotif := b.Client.UpdateApplicationNotifications(app.Notifications, app.Name, "")
		if errNotif != nil {
			b.Logger.Errorf("Failed to update notifications: (%s)", errNotif.Error())
		}
	}

	ids, _ := b.PipelineIDs(app.Name)
	ignoreList := make(map[string]bool)
	idToName := make(map[string]string)
	for name, id := range ids {
		ignoreList[id] = false
		idToName[id] = name
	}
	b.Logger.Infof("Found pipelines for %v: %v", app, ids)
	for _, p := range pipelines {
		// Add ids to existing pipelines
		b.Logger.Info("Processing pipeline ", p)
		if id, exists := ids[p.Name]; exists {
			b.Logger.Debug("Added id ", id, " to pipeline ", p.Name)
			ignoreList[p.Name] = true
			p.ID = id //note: we're working with a copy.  once this loop exits all changes go out of scope!
			b.Logger.Info("Updating pipeline: " + p.Name)
		} else {
			b.Logger.Debug("Adding ", p.Name, " to ignored stale pipelines")
			ignoreList[p.Name] = true
			b.Logger.Info("Creating pipeline: " + p.Name)
		}
		if autoLock == "true" {
			b.Logger.Debug("Locking pipeline ", p.Name)
			p.Lock()
		}

		if err := b.Client.UpsertPipeline(p, p.ID, ""); err != nil {
			err = unwrapFront50Error(err)
			b.Logger.Errorf("Upsert failed: %s", err.Error())
			return err
		}
		b.Logger.Info("Upsert succeeded.")
	}
	if deleteStale {
		// clear existing pipelines that weren't updated
		b.Logger.Debug("Pipelines we should ignore because they were just created: ", ignoreList)
		allPipelines, err := b.Client.GetPipelines(app.Name, "")
		if err != nil {
			b.Logger.Errorf("Could not retrieve pipelines for %s: %s", app.Name, err.Error())
		} else {
			for _, p := range allPipelines {
				if !ignoreList[p.Name] {
					b.Logger.Infof("Deleting stale pipeline %s", p.Name)
					if err := b.Client.DeletePipeline(p, ""); err != nil {
						// Not worrying about handling errors here because it just means it
						// didn't get deleted *this time*.
						b.Logger.Warnf("Could not delete Pipeline %s (Application %s)", p.Name, p.Application)
					}
				}
			}
		}
	}
	return err
}

// PipelineIDs returns a map of pipeline names -> their UUID.
func (b *PipelineBuilder) PipelineIDs(app string) (map[string]string, error) {
	ids := map[string]string{}
	b.Logger.Info("Looking up existing pipelines")
	pipelines, err := b.Client.GetPipelines(app, "")
	if err != nil {
		b.Logger.Errorf("Failed to GetPipelines for %s: %s", app, err.Error())
		return ids, err
	}
	for _, p := range pipelines {
		ids[p.Name] = p.ID
	}
	return ids, nil
}

// GetPipelineByID returns a pipeline's UUID by its name; if the pipeline
// isn't found, one is created one and its ID is returned.
func (b *PipelineBuilder) GetPipelineByID(app, pipelineName string) (string, error) {
	ids, err := b.PipelineIDs(app)
	if err != nil {
		return "", err
	}
	id, exists := ids[pipelineName]
	if exists {
		return id, nil
	}
	err = b.Client.UpsertPipeline(plank.Pipeline{
		Application: app,
		Name:        pipelineName,
	}, "", "")
	if err != nil {
		err = unwrapFront50Error(err)
		b.Logger.Errorf("Failed to UpsertPipeline for %s (%s): %s", pipelineName, app, err.Error())
		return "", err
	}
	return b.GetPipelineByID(app, pipelineName)
}

func (b *PipelineBuilder) AddUnmarshaller(u Unmarshaller) {
	b.Ums = append(b.Ums, u)
}

func (b *PipelineBuilder) NotifySuccess(org, repo, path string, notifications plank.NotificationsType) {
	for _, n := range b.Notifiers {
		if b.Action == pipebuilder.Validate {
			if n.SendOnValidation() {
				n.SendSuccess(org, repo, path, notifications, b.getNotificationContent())
			}
		} else {
			n.SendSuccess(org, repo, path, notifications, b.getNotificationContent())
		}
	}
}

func (b *PipelineBuilder) NotifyFailure(org, repo, path string, err error, dinghyfile string) {
	var notifications plank.NotificationsType
	if appName, err := extractApplicationName(dinghyfile); err == nil {
		if foundNotifications, errGetApp := b.Client.GetApplicationNotifications(appName, ""); errGetApp == nil {
			notifications = *foundNotifications
		}
	}
	for _, n := range b.Notifiers {
		if b.Action == pipebuilder.Validate {
			if n.SendOnValidation() {
				n.SendFailure(org, repo, path, err, notifications, b.getNotificationContent())
			}
		} else {
			n.SendFailure(org, repo, path, err, notifications, b.getNotificationContent())
		}
	}
}

func extractApplicationName(dinghyfile string) (string, error) {
	// Groups name of the application, valid values are application: appname  "application":"appname" and 'application': 'appname'
	regex := `["']?application["']?\s*[:=]\s*["']?(?P<applicationName>[\w-]*)["']?`
	params := getParams(regex, dinghyfile)
	if val, ok := params["applicationName"]; ok {
		return val, nil
	}
	return "", errors.New("application name not found in dinghyfile")
}

// This function returns a map of strings for matches
func getParams(regEx, url string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func (b *PipelineBuilder) getNotificationContent() map[string]interface{} {
	content := map[string]interface{}{}

	logEvent, err := b.Logger.GetBytesBuffByLoggerKey(log.LogEventKey)
	if err != nil {
		b.Logger.Error(fmt.Sprintf("there was an error trying to get notification content: %v", err))
	} else {
		content["logevent"] = logEvent.String()
	}
	if b.PushRaw != nil {
		content["rawdata"] = b.PushRaw
	}
	return content
}
