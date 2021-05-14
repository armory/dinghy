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
	"github.com/armory/dinghy/pkg/dinghyfile/pipebuilder"
	dinghylog "github.com/armory/dinghy/pkg/log"
	"github.com/armory/dinghy/pkg/logevents"
	"github.com/armory/dinghy/pkg/settings/global"
	"github.com/armory/dinghy/pkg/settings/source"
	"io/ioutil"
	"net/http"
	"path/filepath"
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
	"github.com/armory/dinghy/pkg/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	SetCommitStatus(instanceId string, s git.Status, description string)
	GetCommitStatus() (error, git.Status, string)
	GetCommits() []string
	Name() string
}

type MetricsHandler interface {
	WrapHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request))
}

type NoOpMetricsHandler struct{}

func (nom *NoOpMetricsHandler) WrapHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) (string, func(http.ResponseWriter, *http.Request)) {
	return pattern, handler
}

type WebAPI struct {
	SourceConfig    source.SourceConfiguration
	ClientReadOnly  util.PlankClient
	Cache           dinghyfile.DependencyManager
	CacheReadOnly   dinghyfile.DependencyManager
	EventClient     *events.Client
	Logger          log.FieldLogger
	Ums             []dinghyfile.Unmarshaller
	Notifiers       []notifiers.Notifier
	Parser          dinghyfile.Parser
	LogEventsClient logevents.LogEventsClient
	MuxRouter       *mux.Router
	Logr            *log.Logger
	MetricsHandler
}

func NewWebAPI(s source.SourceConfiguration, r dinghyfile.DependencyManager, e *events.Client, l log.FieldLogger, depreadonly dinghyfile.DependencyManager, clientreadonly util.PlankClient, logeventsClient logevents.LogEventsClient, logr *log.Logger) *WebAPI {
	return &WebAPI{
		SourceConfig:    s,
		Cache:           r,
		EventClient:     e,
		Logger:          l,
		Ums:             []dinghyfile.Unmarshaller{},
		Notifiers:       []notifiers.Notifier{},
		ClientReadOnly:  clientreadonly,
		CacheReadOnly:   depreadonly,
		LogEventsClient: logeventsClient,
		Logr:            logr,
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
func (wa *WebAPI) Router(ts TraceSettings) *mux.Router {
	r := mux.NewRouter()
	r.Use(ts.TraceExtract())
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/", wa.healthcheck))
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/health", wa.healthcheck))
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/healthcheck", wa.healthcheck))
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/v1/logevents", wa.logevents)).Methods("GET")
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/v1/webhooks/github", wa.githubWebhookHandler)).Methods("POST")
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/v1/webhooks/gitlab", wa.gitlabWebhookHandler)).Methods("POST")
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/v1/webhooks/stash", wa.stashWebhookHandler)).Methods("POST")
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/v1/webhooks/bitbucket", wa.bitbucketWebhookHandler)).Methods("POST")
	// all of the bitbucket webhooks come through this one handler, this is being left for backwards compatibility
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/v1/webhooks/bitbucket-cloud", wa.bitbucketWebhookHandler)).Methods("POST")
	r.HandleFunc(wa.MetricsHandler.WrapHandleFunc("/v1/updatePipeline", wa.manualUpdateHandler)).Methods("POST")
	r.Use(RequestLoggingMiddleware)
	return r
}

// ==============
// route handlers
// ==============

func (wa *WebAPI) logevents(w http.ResponseWriter, r *http.Request) {
	logEvents, err := wa.LogEventsClient.GetLogEvents()
	if err == nil {

	}
	bytesResult, _ := json.Marshal(logEvents)
	w.Write(bytesResult)
}

func (wa *WebAPI) healthcheck(w http.ResponseWriter, r *http.Request) {
	wa.Logger.Debug(r.RemoteAddr, " Requested ", r.RequestURI)
	w.Write([]byte(`{"status":"ok"}`))
}

func (wa *WebAPI) manualUpdateHandler(w http.ResponseWriter, r *http.Request) {
	logger := DecorateLogger(wa.Logger, RequestContextFields(r.Context()))
	dinghyLog := dinghylog.NewDinghyLogs(logger)
	settings, plankClient, err := wa.SourceConfig.GetSettings(r, wa.Logr)
	if err != nil {
		dinghyLog.Errorf("Failed to get the settings: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	var fileService = dummy.FileService{}

	builder := &dinghyfile.PipelineBuilder{
		Depman:               cache.NewMemoryCache(),
		Downloader:           fileService,
		Client:               plankClient,
		DeleteStalePipelines: false,
		AutolockPipelines:    settings.AutoLockPipelines,
		Logger:               dinghyLog,
		Ums:                  wa.Ums,
		Action:               pipebuilder.Process,
	}

	builder.Parser = wa.Parser
	builder.Parser.SetBuilder(builder)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	fileService["master"] = make(map[string]string)
	fileService["master"]["dinghyfile"] = buf.String()
	wa.Logger.Infof("Received payload: %s", fileService["master"]["dinghyfile"])

	if _, err := builder.ProcessDinghyfile("", "", "dinghyfile", ""); err != nil {
		util.WriteHTTPError(w, http.StatusInternalServerError, err)
	}
}

func (wa *WebAPI) githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	logger := DecorateLogger(wa.Logger, RequestContextFields(r.Context()))
	dinghyLog := dinghylog.NewDinghyLogs(logger)
	settings, plankClient, err := wa.SourceConfig.GetSettings(r, wa.Logr)
	if err != nil {
		dinghyLog.Errorf("Failed to get the settings: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	p := github.Push{Logger: dinghyLog}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		dinghyLog.Errorf("failed to read body in github webhook handler: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	dinghyLog.Infof("Received payload: %s", string(body))
	if err := json.Unmarshal(body, &p); err != nil {
		dinghyLog.Errorf("failed to decode github webhook: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if p.Ref == "" {
		// Unmarshal failed, might be a non-Push notification. Log event and return
		dinghyLog.Info("Possibly a non-Push notification received (blank ref)")
		return
	}

	provider := "github"
	enabled := contains(settings.WebhookValidationEnabledProviders, provider)

	if enabled {
		repo := p.Repo()
		org := p.Org()
		whvalidations := settings.WebhookValidations
		if whvalidations != nil && len(whvalidations) > 0 {
			if !validateWebhookSignature(whvalidations, repo, org, provider, body, r, dinghyLog) {
				saveLogEventError(wa.LogEventsClient, &p, dinghyLog, logevents.LogEvent{RawData: string(body)})
				return
			}
		}
	}

	p.Ref = strings.Replace(p.Ref, "refs/heads/", "", 1)

	// TODO: we're assigning config in two places here, we should refactor this
	gh := github.Config{Endpoint: settings.GithubEndpoint, Token: settings.GitHubToken}
	p.Config = gh
	p.DeckBaseURL = settings.Deck.BaseURL
	fileService := github.FileService{GitHub: &gh, Logger: dinghyLog}

	var pullRequestUrl string
	if pullRequest, err := gh.GetPullRequest(p.Org(), p.Repo(), p.Branch(), gh.GetShaFromRawData(body)); err == nil {
		if pullRequest != nil {
			pullRequestUrl = pullRequest.GetHTMLURL()
		}
	}

	wa.buildPipelines(&p, body, &fileService, w, dinghyLog, pullRequestUrl, plankClient, settings)
}

func contains(whvalidations []string, provider string) bool {
	if whvalidations == nil {
		return false
	}
	for _, val := range whvalidations {
		if val == provider {
			return true
		}
	}
	return false
}

func validateWebhookSignature(whvalidations []global.WebhookValidation, repo string, org string, provider string, body []byte, r *http.Request, logger dinghylog.DinghyLog) bool {
	whcurrentvalidation := global.WebhookValidation{}
	if found, whval := findWebhookValidation(whvalidations, repo, org, provider); found {
		//If record is found and validation is disabled then just return true
		if whval.Enabled == false {
			logger.Infof("Webhook validation for %v/%v is disabled so validation will by bypassed", org, repo)
			return true
		}
		whcurrentvalidation = *whval
	} else {
		logger.Infof("Webhook validation for %v/%v was not found, searching for default-webhook-secret", org, repo)
		if foundDefault, whvalDefault := findWebhookValidation(whvalidations, "default-webhook-secret", org, provider); foundDefault {
			if whvalDefault.Enabled == true {
				whcurrentvalidation = *whvalDefault
				logger.Infof("Webhook default secret was found for org: %v", org)
			} else {
				logger.Infof("Webhook default secret for org: %v is disabled", org)
				return false
			}
		} else {
			logger.Infof("Webhook validation for %v/%v and default-webhook-secret were not found", org, repo)
			return false
		}
	}
	rawPayload := getRawPayload(body)
	//X-Hub-Signature is the original header from github, but since this message is from echo we receive webhook-secret
	whsecret := getHeader(r, "webhook-secret")

	if rawPayload == "" || whsecret == "" {
		//Validate in webhook and raw_payload and webhook secret is present.
		logger.Error("There is a webhook validation registered in dinghy but the webhook is not configured in github side")
		return false
	}

	return github.IsValidSignature([]byte(rawPayload), whsecret, whcurrentvalidation.Secret, logger)
}

func findWebhookValidation(whvalidations []global.WebhookValidation, repo string, org string, provider string) (bool, *global.WebhookValidation) {
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

	attributeValue := "raw_payload"
	if m[attributeValue] != nil {
		return fmt.Sprintf("%v", m[attributeValue])
	}
	return ""
}

func getHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func (wa *WebAPI) gitlabWebhookHandler(w http.ResponseWriter, r *http.Request) {
	logger := DecorateLogger(wa.Logger, RequestContextFields(r.Context()))
	dinghyLog := dinghylog.NewDinghyLogs(logger)
	settings, plankClient, err := wa.SourceConfig.GetSettings(r, wa.Logr)
	if err != nil {
		dinghyLog.Errorf("Failed to get the settings: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	p := gitlab.Push{Logger: dinghyLog}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		dinghyLog.Errorf("failed to read body in gitlab webhook handler: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	dinghyLog.Infof("Received payload: %s", string(body))

	fileService, err := p.ParseWebhook(settings, body)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected event type") {
			dinghyLog.Infof("Non-Push gitlab notification (%s)", strings.SplitN(err.Error(), ":", 2))
			saveLogEventError(wa.LogEventsClient, &p, dinghyLog, logevents.LogEvent{RawData: string(body)})
			return
		}
		dinghyLog.Errorf("failed to parse gitlab webhook: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		saveLogEventError(wa.LogEventsClient, &p, dinghyLog, logevents.LogEvent{RawData: string(body)})
		return
	}
	wa.buildPipelines(&p, body, &fileService, w, dinghyLog, "", plankClient, settings)
}

func (wa *WebAPI) stashWebhookHandler(w http.ResponseWriter, r *http.Request) {
	logger := DecorateLogger(wa.Logger, RequestContextFields(r.Context()))
	dinghyLog := dinghylog.NewDinghyLogs(logger)
	settings, plankClient, err := wa.SourceConfig.GetSettings(r, wa.Logr)
	if err != nil {
		dinghyLog.Errorf("Failed to get the settings: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	payload := stash.WebhookPayload{}

	dinghyLog.Infof("Reading stash payload body")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		dinghyLog.Errorf("failed to read body in stash webhook handler: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer r.Body.Close()
	dinghyLog.Infof("Received payload: %s", string(body))
	if err := json.Unmarshal(body, &payload); err != nil {
		dinghyLog.Errorf("failed to decode stash webhook: %s", err.Error())
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	payload.IsOldStash = true
	stashConfig := stash.Config{
		Endpoint: settings.StashEndpoint,
		Username: settings.StashUsername,
		Token:    settings.StashToken,
		Logger:   dinghyLog,
	}
	dinghyLog.Infof("Instantiating Stash Payload")
	p, err := stash.NewPush(payload, stashConfig)
	if err != nil {
		dinghyLog.Warnf("stash.NewPush failed: %s", err.Error())
		util.WriteHTTPError(w, http.StatusInternalServerError, err)
		saveLogEventError(wa.LogEventsClient, p, dinghyLog, logevents.LogEvent{RawData: string(body)})
		return
	}

	// TODO: WebAPI already has the fields that are being assigned here and it's
	// the receiver on the buildPipelines. We don't need to reassign the values to
	// fileService here.
	fileService := stash.FileService{
		Config: stashConfig,
		Logger: dinghyLog,
	}
	dinghyLog.Infof("Building pipeslines from Stash webhook")
	wa.buildPipelines(p, body, &fileService, w, dinghyLog, "", plankClient, settings)
}

func (wa *WebAPI) bitbucketWebhookHandler(w http.ResponseWriter, r *http.Request) {
	logger := DecorateLogger(wa.Logger, RequestContextFields(r.Context()))
	dinghyLog := dinghylog.NewDinghyLogs(logger)
	settings, plankClient, err := wa.SourceConfig.GetSettings(r, wa.Logr)
	if err != nil {
		dinghyLog.Errorf("Failed to get the settings: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	// read the response body to check for the type and use NopCloser so it can be decoded later
	keys := make(map[string]interface{})
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		dinghyLog.Errorf("Failed to read request body: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer r.Body.Close()

	r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	if err := json.Unmarshal(b, &keys); err != nil {
		dinghyLog.Errorf("Unable to determine bitbucket event type: %s", err)
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Some bitbucket versions do not send the  X-Event-Key this causes event_type to not be parsed
	// We will take the value eventKey instead since is the same
	if _, found := keys["event_type"]; !found {
		dinghyLog.Info("event_type was not found in payload, so eventKey will be used instead")
		keys["event_type"] = keys["eventKey"]
	}

	switch keys["event_type"] {
	case "repo:push", "pullrequest:fulfilled":
		dinghyLog.Info("Processing bitbucket-cloud webhook")
		payload := bbcloud.WebhookPayload{}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			dinghyLog.Errorf("failed to read body in bitbucket-cloud webhook handler: %s", err.Error())
			util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
			return
		}
		defer r.Body.Close()
		dinghyLog.Infof("Received payload: %s", string(body))
		dinghyLog.Infof("Received payload: %s", string(body))
		if err := json.Unmarshal(body, &payload); err != nil {
			dinghyLog.Errorf("failed to decode bitbucket-cloud webhook: %s", err.Error())
			util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
			return
		}

		bbcloudConfig := bbcloud.Config{
			Endpoint: settings.StashEndpoint,
			Username: settings.StashUsername,
			Token:    settings.StashToken,
			Logger:   dinghyLog,
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
			Logger: dinghyLog,
		}

		wa.buildPipelines(p, body, &fileService, w, dinghyLog, "", plankClient, settings)

	case "repo:refs_changed", "pr:merged":
		dinghyLog.Info("Processing bitbucket-server webhook")
		payload := stash.WebhookPayload{}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			dinghyLog.Errorf("failed to read body in bitbucket-server webhook handler: %s", err.Error())
			util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
			return
		}
		defer r.Body.Close()
		dinghyLog.Infof("Received payload: %s", string(body))
		if err := json.Unmarshal(body, &payload); err != nil {
			dinghyLog.Errorf("failed to decode bitbucket-server webhook: %s", err.Error())
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
			Endpoint: settings.StashEndpoint,
			Username: settings.StashUsername,
			Token:    settings.StashToken,
			Logger:   dinghyLog,
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
			Logger: dinghyLog,
		}

		wa.buildPipelines(p, body, &fileService, w, dinghyLog, "", plankClient, settings)

	default:
		util.WriteHTTPError(w, http.StatusInternalServerError, errors.New("Unknown bitbucket event type"))
		return
	}
}

// =========
// utilities
// =========

// ProcessPush processes a push using a pipeline builder
func (wa *WebAPI) ProcessPush(p Push, b *dinghyfile.PipelineBuilder, settings *global.Settings) (string, error) {
	// Ensure dinghyfile was changed.
	if !p.ContainsFile(settings.DinghyFilename) {
		b.Logger.Infof("Push does not include %s, skipping.", settings.DinghyFilename)
		errstat, status, _ := p.GetCommitStatus()
		if errstat == nil && status == "" {
			p.SetCommitStatus(settings.InstanceId, git.StatusSuccess, fmt.Sprintf("No changes in %v.", settings.DinghyFilename))
		}
		return "", nil
	}

	b.Logger.Info("Dinghyfile found in commit for repo " + p.Repo())

	// Set commit status to the pending yellow dot.
	p.SetCommitStatus(settings.InstanceId, git.StatusPending, git.DefaultMessagesByBuilderAction[b.Action][git.StatusPending])

	var dinghyfilesRendered bytes.Buffer
	for _, filePath := range p.Files() {
		components := strings.Split(filePath, "/")
		if components[len(components)-1] == settings.DinghyFilename {
			// Process the dinghyfile.
			dinghyRendered, err := b.ProcessDinghyfile(p.Org(), p.Repo(), filePath, p.Branch())
			dinghyfilesRendered.WriteString(dinghyRendered)
			// Set commit status based on result of processing.
			if err != nil {
				if err == dinghyfile.ErrMalformedJSON {
					b.Logger.Errorf("Error processing Dinghyfile (malformed JSON): %s", err.Error())
					p.SetCommitStatus(settings.InstanceId, git.StatusFailure, "Error processing Dinghyfile (malformed JSON)")
				} else {
					b.Logger.Errorf("Error processing Dinghyfile: %s", err.Error())
					p.SetCommitStatus(settings.InstanceId, git.StatusError, fmt.Sprintf("%s", err.Error()))
				}
				return dinghyfilesRendered.String(), err
			}
			p.SetCommitStatus(settings.InstanceId, git.StatusSuccess, git.DefaultMessagesByBuilderAction[b.Action][git.StatusSuccess])
		}
	}
	return dinghyfilesRendered.String(), nil
}

// TODO: this func should return an error and allow the handlers to return the http response. Additionally,
// it probably doesn't belong in this file once refactored.
func (wa *WebAPI) buildPipelines(p Push, rawPush []byte, f dinghyfile.Downloader, w http.ResponseWriter, dinghyLog dinghylog.DinghyLog, pullRequest string, plankClient util.PlankClient, settings *global.Settings) {
	// see if we have any configurations for this repo.
	// if we do have configurations, see if this is the branch we want to use. If it's not, skip and return.
	var validation bool
	if rc := settings.GetRepoConfig(p.Name(), p.Repo()); rc != nil {
		if !p.IsBranch(rc.Branch) {
			dinghyLog.Infof("Received request from branch %s. Does not match configured branch %s. Proceeding as validation.", p.Branch(), rc.Branch)
			validation = true
		}
	} else {
		// if we didn't find any configurations for this repo, proceed with master
		dinghyLog.Infof("Found no custom configuration for repo: %s, proceeding with master", p.Repo())
		if !p.IsMaster() {
			dinghyLog.Infof("Skipping Spinnaker pipeline update because this branch (%s) is not master. Proceeding as validation.", p.Branch())
			validation = true
		}
	}

	dinghyLog.Infof("Processing request for branch: %s", p.Branch())

	// deserialze push data to a map.  used in template logic later
	rawPushData := make(map[string]interface{})
	if err := json.Unmarshal(rawPush, &rawPushData); err != nil {
		dinghyLog.Errorf("unable to deserialze raw data to map")
	}

	// Construct a pipeline builder using provided downloader
	builder := &dinghyfile.PipelineBuilder{
		Downloader:                  f,
		Depman:                      wa.Cache,
		TemplateRepo:                settings.TemplateRepo,
		TemplateOrg:                 settings.TemplateOrg,
		DinghyfileName:              settings.DinghyFilename,
		DeleteStalePipelines:        false,
		AutolockPipelines:           settings.AutoLockPipelines,
		Client:                      plankClient,
		EventClient:                 wa.EventClient,
		Logger:                      dinghyLog,
		Ums:                         wa.Ums,
		Notifiers:                   wa.Notifiers,
		PushRaw:                     rawPushData,
		RepositoryRawdataProcessing: settings.RepositoryRawdataProcessing,
		Action:                      pipebuilder.Process,
	}

	if validation {
		builder.Client = wa.ClientReadOnly
		builder.Depman = wa.CacheReadOnly
		builder.Action = pipebuilder.Validate
	}

	builder.Parser = wa.Parser
	builder.Parser.SetBuilder(builder)

	// Process the push.
	dinghyLog.Info("Processing Push")
	dinghyfilesRendered, err := wa.ProcessPush(p, builder, settings)
	if err == dinghyfile.ErrMalformedJSON {
		util.WriteHTTPError(w, http.StatusUnprocessableEntity, err)
		dinghyLog.Errorf("ProcessPush Failed (malformed JSON): %s", err.Error())
		saveLogEventError(wa.LogEventsClient, p, dinghyLog, logevents.LogEvent{RawData: string(rawPush), PullRequest: pullRequest, RenderedDinghyfile: dinghyfilesRendered})
		return
	} else if err != nil {
		dinghyLog.Errorf("ProcessPush Failed (other): %s", err.Error())
		util.WriteHTTPError(w, http.StatusInternalServerError, err)
		saveLogEventError(wa.LogEventsClient, p, dinghyLog, logevents.LogEvent{RawData: string(rawPush), PullRequest: pullRequest, RenderedDinghyfile: dinghyfilesRendered})
		return
	}

	// Check if we're in a template repo
	if p.Repo() == settings.TemplateRepo {
		// Set status to pending while we process modules
		p.SetCommitStatus(settings.InstanceId, git.StatusPending, git.DefaultMessagesByBuilderAction[builder.Action][git.StatusPending])

		// For each module pushed, rebuild dependent dinghyfiles
		for _, file := range p.Files() {
			// ensure module is correctly parsed
			if _, err := builder.Parser.Parse(p.Org(), p.Repo(), file, p.Branch(), nil); err != nil {
				p.SetCommitStatus(settings.InstanceId, git.StatusError, "module parse failed")
				dinghyLog.Errorf("module parse failed: %s", err.Error())
				saveLogEventError(wa.LogEventsClient, p, dinghyLog, logevents.LogEvent{RawData: string(rawPush), PullRequest: pullRequest, RenderedDinghyfile: dinghyfilesRendered})
				return
			}
			if err := builder.RebuildModuleRoots(p.Org(), p.Repo(), file, p.Branch()); err != nil {
				switch err.(type) {
				case *util.GitHubFileNotFoundErr:
					util.WriteHTTPError(w, http.StatusNotFound, err)
				default:
					util.WriteHTTPError(w, http.StatusInternalServerError, err)
				}
				p.SetCommitStatus(settings.InstanceId, git.StatusError, "Rebuilding dependent dinghyfiles Failed")
				dinghyLog.Errorf("RebuildModuleRoots Failed: %s", err.Error())
				saveLogEventError(wa.LogEventsClient, p, dinghyLog, logevents.LogEvent{RawData: string(rawPush), PullRequest: pullRequest, RenderedDinghyfile: dinghyfilesRendered})
				return
			}
		}
		p.SetCommitStatus(settings.InstanceId, git.StatusSuccess, git.DefaultMessagesByBuilderAction[builder.Action][git.StatusSuccess])
	}

	// Only save event if changed files were in repo or it was having a dinghyfile
	// TODO: If a template repo is having files not related with dinghy an event will be saved
	if p.Repo() == settings.TemplateRepo {
		if len(p.Files()) > 0 {
			saveLogEventSuccess(wa.LogEventsClient, p, dinghyLog, logevents.LogEvent{RawData: string(rawPush), PullRequest: pullRequest, RenderedDinghyfile: dinghyfilesRendered})
		}
	} else {
		dinghyfiles := []string{}
		for _, currfile := range p.Files() {
			if filepath.Base(currfile) == builder.DinghyfileName {
				dinghyfiles = append(dinghyfiles, currfile)
			}
		}
		if len(dinghyfiles) > 0 {
			saveLogEventSuccess(wa.LogEventsClient, p, dinghyLog, logevents.LogEvent{RawData: string(rawPush), Files: dinghyfiles, PullRequest: pullRequest, RenderedDinghyfile: dinghyfilesRendered})
		}
	}
	w.Write([]byte(`{"status":"accepted"}`))
}
