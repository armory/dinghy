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

package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/armory/dinghy/pkg/events"
	"github.com/armory/dinghy/pkg/git/bbcloud"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/armory/dinghy/pkg/cache"
	"github.com/armory/dinghy/pkg/dinghyfile"
	"github.com/armory/dinghy/pkg/git"
	"github.com/armory/dinghy/pkg/git/dummy"
	"github.com/armory/dinghy/pkg/git/github"
	"github.com/armory/dinghy/pkg/git/gitlab"
	"github.com/armory/dinghy/pkg/git/stash"
	"github.com/armory/dinghy/pkg/notifiers"
	"github.com/armory/dinghy/pkg/settings"
	"github.com/armory/dinghy/pkg/util"
)

// Push represents a push notification from a git service.
type Push interface {
	ContainsFile(file string) bool
	Files() []string
	Repo() string
	Org() string
	Branch() string
	IsBranch(string) bool
	IsMaster() bool
	SetCommitStatus(s git.Status, description string)
	Name() string
}

type WebAPI struct {
	Config      *settings.Settings
	Client      util.PlankClient
	Cache       dinghyfile.DependencyManager
	EventClient *events.Client
	Logger      log.FieldLogger
	Ums         []dinghyfile.Unmarshaller
	Notifiers   []notifiers.Notifier
	Parser      dinghyfile.Parser
}

func NewWebAPI(s *settings.Settings, r dinghyfile.DependencyManager, c util.PlankClient, e *events.Client, l log.FieldLogger) *WebAPI {
	return &WebAPI{
		Config:      s,
		Client:      c,
		Cache:       r,
		EventClient: e,
		Logger:      l,
		Ums:         []dinghyfile.Unmarshaller{},
		Notifiers:   []notifiers.Notifier{},
	}
}

func (wa *WebAPI) AddDinghyfileUnmarshaller(u dinghyfile.Unmarshaller) {
	wa.Ums = append(wa.Ums, u)
}

func (wa *WebAPI) SetDinghyfileParser(p dinghyfile.Parser) {
	wa.Parser = p
}

// AddNotifier adds a Notifier type instance that will be triggered when
// a Dinghyfile processing phase completes (success/fail).  It only gets
// triggered if there is work to do on a push (ie. a pipeline is intended
// to be updated)
func (wa *WebAPI) AddNotifier(n notifiers.Notifier) {
	wa.Notifiers = append(wa.Notifiers, n)
}

// Router defines the routes for the application.
func (wa *WebAPI) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", wa.healthcheck)
	r.HandleFunc("/health", wa.healthcheck)
	r.HandleFunc("/healthcheck", wa.healthcheck)
	r.HandleFunc("/v1/webhooks/github", wa.githubWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/webhooks/gitlab", wa.gitlabWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/webhooks/stash", wa.stashWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/webhooks/bitbucket", wa.bitbucketWebhookHandler).Methods("POST")
	// all of the bitbucket webhooks come through this one handler, this is being left for backwards compatibility
	r.HandleFunc("/v1/webhooks/bitbucket-cloud", wa.bitbucketWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/updatePipeline", wa.manualUpdateHandler).Methods("POST")
	r.Use(RequestLoggingMiddleware)
	return r
}

// ==============
// route handlers
// ==============

func (wa *WebAPI) healthcheck(w http.ResponseWriter, r *http.Request) {
	wa.Logger.Debug(r.RemoteAddr, " Requested ", r.RequestURI)
	w.Write([]byte(`{"status":"ok"}`))
}

func (wa *WebAPI) manualUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var fileService = dummy.FileService{}

	builder := &dinghyfile.PipelineBuilder{
		Depman:               cache.NewMemoryCache(),
		Downloader:           fileService,
		Client:               wa.Client,
		DeleteStalePipelines: false,
		AutolockPipelines:    wa.Config.AutoLockPipelines,
		Logger:               wa.Logger,
		Ums:                  wa.Ums,
	}

	builder.Parser = wa.Parser
	builder.Parser.SetBuilder(builder)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	fileService["master"] = make(map[string]string)
	fileService["master"]["dinghyfile"] = buf.String()
	wa.Logger.Infof("Received payload: %s", fileService["master"]["dinghyfile"])

	if err := builder.ProcessDinghyfile("", "", "dinghyfile", ""); err != nil {
		util.WriteHTTPError(w, http.StatusInternalServerError, err)
	}
}

func (wa *WebAPI) githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	p := github.Push{Logger: wa.Logger}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		wa.Logger.Errorf("failed to read body in github webhook handler: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	wa.Logger.Infof("Received payload: %s", string(body))
	if err := json.Unmarshal(body, &p); err != nil {
		wa.Logger.Errorf("failed to decode github webhook: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if p.Ref == "" {
		// Unmarshal failed, might be a non-Push notification. Log event and return
		wa.Logger.Info("Possibly a non-Push notification received (blank ref)")
		return
	}

	provider := "github"
	enabled := contains(wa.Config.WebhookValidationEnabledProviders, provider)

	if enabled {
		repo := p.Repository.Name
		org := p.Repository.Organization
		whvalidations := wa.Config.WebhookValidations
		if whvalidations != nil && len(whvalidations) > 0 {
			if !validateWebhookSignature(whvalidations, repo, org, provider, body, r, wa) {
				return
			}
		}
	}

	p.Ref = strings.Replace(p.Ref, "refs/heads/", "", 1)

	// TODO: we're assigning config in two places here, we should refactor this
	gh := github.Config{Endpoint: wa.Config.GithubEndpoint, Token: wa.Config.GitHubToken}
	p.Config = gh
	p.DeckBaseURL = wa.Config.Deck.BaseURL
	fileService := github.FileService{GitHub: &gh, Logger: wa.Logger}

	wa.buildPipelines(&p, body, &fileService, w)
}

func contains(whvalidations []string, provider string) bool {
	if whvalidations == nil {
		return false
	}
	for _, val := range whvalidations  {
		if val == provider {
			return true
		}
	}
	return false
}

func validateWebhookSignature(whvalidations []settings.WebhookValidation, repo string, org string, provider string, body []byte, r *http.Request, wa *WebAPI) bool {
	whcurrentvalidation := settings.WebhookValidation{}
	if found,whval := findWebhookValidation(whvalidations, repo, org, provider); found {
		//If record is found and validation is disabled then just return true
		if whval.Enabled == false {
			wa.Logger.Infof("Webhook validation for %v/%v is disabled so validation will by bypassed", org, repo)
			return true
		}
		whcurrentvalidation = *whval
	} else {
		wa.Logger.Infof("Webhook validation for %v/%v was not found, searching for default-webhook-secret", org, repo)
		if foundDefault, whvalDefault := findWebhookValidation(whvalidations, "default-webhook-secret", org, provider); foundDefault{
			if whvalDefault.Enabled == true {
				whcurrentvalidation = *whvalDefault
				wa.Logger.Infof("Webhook default secret was found for org: %v", org)
			} else {
				wa.Logger.Infof("Webhook default secret for org: %v is disabled", org)
				return false
			}
		} else {
			wa.Logger.Infof("Webhook validation for %v/%v and default-webhook-secret were not found", org, repo)
			return false
		}
	}
	rawPayload := getRawPayload(body)
	whsecret := getWebhookSecret(r)

	if rawPayload == "" || whsecret == "" {
		//Validate in webhook and raw_payload and webhook secret is present.
		wa.Logger.Error("There is a webhook validation registered in dinghy but the webhook is not configured in github side")
		return false
	}

	return github.IsValidSignature([]byte(rawPayload), getWebhookSecret(r), whcurrentvalidation.Secret, wa.Logger)
}

func findWebhookValidation(whvalidations []settings.WebhookValidation, repo string, org string, provider string) (bool, *settings.WebhookValidation) {
	if whvalidations != nil && len(whvalidations) > 0 {
		for i := range whvalidations {
			whval := whvalidations[i]
			if whval.Repo == repo && whval.Organization == org && whval.VersionControlProvider == provider {
				return true, &whval
			}
		}
	}
	return false, nil
}

func getRawPayload(body []byte) string {
	m := make(map[string]interface{})
	err := json.Unmarshal(body, &m)
	if err != nil {
		log.Error("Failed to parse body json.")
	}

	attributeValue :="raw_payload"
	if m[attributeValue] != nil {
		return fmt.Sprintf("%v",m[attributeValue])
	}
	return ""
}

func getWebhookSecret(r *http.Request) string {
	//X-Hub-Signature is the original header from github, but since this message is from echo we receive webhook-secret
	return r.Header.Get("webhook-secret")
}

func (wa *WebAPI) gitlabWebhookHandler(w http.ResponseWriter, r *http.Request) {
	p := gitlab.Push{Logger: wa.Logger}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		wa.Logger.Errorf("failed to read body in gitlab webhook handler: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	wa.Logger.Infof("Received payload: %s", string(body))

	fileService, err := p.ParseWebhook(wa.Config, body)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected event type") {
			wa.Logger.Infof("Non-Push gitlab notification (%s)", strings.SplitN(err.Error(), ":", 2))
			return
		}
		wa.Logger.Errorf("failed to parse gitlab webhook: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	wa.buildPipelines(&p, body, &fileService, w)
}

func (wa *WebAPI) stashWebhookHandler(w http.ResponseWriter, r *http.Request) {
	payload := stash.WebhookPayload{}

	wa.Logger.Infof("Reading stash payload body")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		wa.Logger.Errorf("failed to read body in stash webhook handler: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer r.Body.Close()
	wa.Logger.Infof("Received payload: %s", string(body))
	if err := json.Unmarshal(body, &payload); err != nil {
		wa.Logger.Errorf("failed to decode stash webhook: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	payload.IsOldStash = true
	stashConfig := stash.Config{
		Endpoint: wa.Config.StashEndpoint,
		Username: wa.Config.StashUsername,
		Token:    wa.Config.StashToken,
		Logger:   wa.Logger,
	}
	wa.Logger.Infof("Instantiating Stash Payload")
	p, err := stash.NewPush(payload, stashConfig)
	if err != nil {
		wa.Logger.Warnf("stash.NewPush failed: %s", err.Error())
		util.WriteHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	// TODO: WebAPI already has the fields that are being assigned here and it's
	// the receiver on the buildPipelines. We don't need to reassign the values to
	// fileService here.
	fileService := stash.FileService{
		Config: stashConfig,
		Logger: wa.Logger,
	}
	wa.Logger.Infof("Building pipeslines from Stash webhook")
	wa.buildPipelines(p, body, &fileService, w)
}

func (wa *WebAPI) bitbucketWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// read the response body to check for the type and use NopCloser so it can be decoded later
	keys := make(map[string]interface{})
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		wa.Logger.Errorf("Failed to read request body: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer r.Body.Close()

	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	if err := json.Unmarshal(b, &keys); err != nil {
		wa.Logger.Errorf("Unable to determine bitbucket event type: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	switch keys["event_type"] {
	case "repo:push", "pullrequest:fulfilled":
		wa.Logger.Info("Processing bitbucket-cloud webhook")
		payload := bbcloud.WebhookPayload{}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			wa.Logger.Errorf("failed to read body in bitbucket-cloud webhook handler: %s", err.Error())
			util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
			return
		}
		defer r.Body.Close()
		wa.Logger.Infof("Received payload: %s", string(body))
		if err := json.Unmarshal(body, &payload); err != nil {
			wa.Logger.Errorf("failed to decode bitbucket-cloud webhook: %s", err.Error())
			util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
			return
		}

		bbcloudConfig := bbcloud.Config{
			Endpoint: wa.Config.StashEndpoint,
			Username: wa.Config.StashUsername,
			Token:    wa.Config.StashToken,
			Logger:   wa.Logger,
		}
		p, err := bbcloud.NewPush(payload, bbcloudConfig)
		if err != nil {
			util.WriteHTTPError(w, http.StatusInternalServerError, err)
			return
		}

		// TODO: WebAPI already has the fields that are being assigned here and it's
		// the receiver on buildPipelines. We don't need to reassign the values to
		// fileService here.
		fileService := bbcloud.FileService{
			Config: bbcloudConfig,
			Logger: wa.Logger,
		}

		wa.buildPipelines(p, body, &fileService, w)

	case "repo:refs_changed", "pr:merged":
		wa.Logger.Info("Processing bitbucket-server webhook")
		payload := stash.WebhookPayload{}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			wa.Logger.Errorf("failed to read body in bitbucket-server webhook handler: %s", err.Error())
			util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
			return
		}
		defer r.Body.Close()
		wa.Logger.Infof("Received payload: %s", string(body))
		if err := json.Unmarshal(body, &payload); err != nil {
			wa.Logger.Errorf("failed to decode bitbucket-server webhook: %s", err.Error())
			util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
			return
		}

		if payload.EventKey != "" && payload.EventKey != "repo:refs_changed" {
			// Not a commit, not an error, we're good.
			w.WriteHeader(200)
			return
		}

		payload.IsOldStash = false
		stashConfig := stash.Config{
			Endpoint: wa.Config.StashEndpoint,
			Username: wa.Config.StashUsername,
			Token:    wa.Config.StashToken,

			Logger: wa.Logger,
		}
		p, err := stash.NewPush(payload, stashConfig)
		if err != nil {
			util.WriteHTTPError(w, http.StatusInternalServerError, err)
			return
		}

		// TODO: WebAPI already has the fields that are being assigned here and it's
		// the receiver on buildPipelines. We don't need to reassign the values to
		// fileService here.
		fileService := stash.FileService{
			Config: stashConfig,
			Logger: wa.Logger,
		}

		wa.buildPipelines(p, body, &fileService, w)

	default:
		util.WriteHTTPError(w, http.StatusInternalServerError, errors.New("Unknown bitbucket event type"))
		return
	}
}

// =========
// utilities
// =========

// ProcessPush processes a push using a pipeline builder
func (wa *WebAPI) ProcessPush(p Push, b *dinghyfile.PipelineBuilder) error {
	// Ensure dinghyfile was changed.
	if !p.ContainsFile(wa.Config.DinghyFilename) {
		wa.Logger.Infof("Push does not include %s, skipping.", wa.Config.DinghyFilename)
		return nil
	}

	wa.Logger.Info("Dinghyfile found in commit for repo " + p.Repo())

	// Set commit status to the pending yellow dot.
	p.SetCommitStatus(git.StatusPending, "")

	for _, filePath := range p.Files() {
		components := strings.Split(filePath, "/")
		if components[len(components)-1] == wa.Config.DinghyFilename {
			// Process the dinghyfile.
			err := b.ProcessDinghyfile(p.Org(), p.Repo(), filePath, p.Branch())
			// Set commit status based on result of processing.
			if err != nil {
				if err == dinghyfile.ErrMalformedJSON {
					wa.Logger.Errorf("Error processing Dinghyfile (malformed JSON): %s", err.Error())
					p.SetCommitStatus(git.StatusFailure, "Error processing Dinghyfile (malformed JSON)" )
				} else {
					wa.Logger.Errorf("Error processing Dinghyfile: %s", err.Error())
					p.SetCommitStatus(git.StatusError, fmt.Sprintf("%s", err.Error()))
				}
				return err
			}
			p.SetCommitStatus(git.StatusSuccess,"")
		}
	}
	return nil
}

// TODO: this func should return an error and allow the handlers to return the http response. Additionally,
// it probably doesn't belong in this file once refactored.
func (wa *WebAPI) buildPipelines(p Push, rawPush []byte, f dinghyfile.Downloader, w http.ResponseWriter) {
	// see if we have any configurations for this repo.
	// if we do have configurations, see if this is the branch we want to use. If it's not, skip and return.
	if rc := wa.Config.GetRepoConfig(p.Name(), p.Repo()); rc != nil {
		if !p.IsBranch(rc.Branch) {
			wa.Logger.Infof("Received request from branch %s. Does not match configured branch %s. Skipping.", p.Branch(), rc.Branch)
			return
		}
	} else {
		// if we didn't find any configurations for this repo, proceed with master
		wa.Logger.Infof("Found no custom configuration for repo: %s, proceeding with master", p.Repo())
		if !p.IsMaster() {
			wa.Logger.Infof("Skipping Spinnaker pipeline update because this branch (%s) is not master", p.Branch())
			return
		}
	}

	wa.Logger.Infof("Processing request for branch: %s", p.Branch())

	// deserialze push data to a map.  used in template logic later
	rawPushData := make(map[string]interface{})
	if err := json.Unmarshal(rawPush, &rawPushData); err != nil {
		wa.Logger.Errorf("unable to deserialze raw data to map")
	}

	// Construct a pipeline builder using provided downloader
	builder := &dinghyfile.PipelineBuilder{
		Downloader:           f,
		Depman:               wa.Cache,
		TemplateRepo:         wa.Config.TemplateRepo,
		TemplateOrg:          wa.Config.TemplateOrg,
		DinghyfileName:       wa.Config.DinghyFilename,
		DeleteStalePipelines: false,
		AutolockPipelines:    wa.Config.AutoLockPipelines,
		Client:               wa.Client,
		EventClient:          wa.EventClient,
		Logger:               wa.Logger,
		Ums:                  wa.Ums,
		Notifiers:            wa.Notifiers,
		PushRaw:              rawPushData,
	}

	builder.Parser = wa.Parser
	builder.Parser.SetBuilder(builder)

	// Process the push.
	wa.Logger.Info("Processing Push")
	err := wa.ProcessPush(p, builder)
	if err == dinghyfile.ErrMalformedJSON {
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		wa.Logger.Errorf("ProcessPush Failed (malformed JSON): %s", err.Error())
		return
	} else if err != nil {
		wa.Logger.Errorf("ProcessPush Failed (other): %s", err.Error())
		util.WriteHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	// Check if we're in a template repo
	if p.Repo() == wa.Config.TemplateRepo {
		// Set status to pending while we process modules
		p.SetCommitStatus(git.StatusPending,"")

		// For each module pushed, rebuild dependent dinghyfiles
		for _, file := range p.Files() {
			if err := builder.RebuildModuleRoots(p.Org(), p.Repo(), file, p.Branch()); err != nil {
				switch err.(type) {
				case *util.GitHubFileNotFoundErr:
					util.WriteHTTPError(w, http.StatusNotFound, err)
				default:
					util.WriteHTTPError(w, http.StatusInternalServerError, err)
				}
				p.SetCommitStatus(git.StatusError,"Rebuilding dependent dinghyfiles Failed")
				wa.Logger.Errorf("RebuildModuleRoots Failed: %s", err.Error())
				return
			}
		}
		p.SetCommitStatus(git.StatusSuccess,"")
	}

	w.Write([]byte(`{"status":"accepted"}`))
}
