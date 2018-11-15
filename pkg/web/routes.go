package web

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/git/dummy"
	"github.com/armory-io/dinghy/pkg/git/github"
	"github.com/armory-io/dinghy/pkg/git/stash"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/util"
)

// Push represents a push notification from a git service.
type Push interface {
	ContainsFile(file string) bool
	Files() []string
	Repo() string
	Org() string
	IsMaster() bool
	SetCommitStatus(s git.Status)
}

// GlobalCache is a global dependency manager to be set on program start
var GlobalCache dinghyfile.DependencyManager

// Router defines the routes for the application.
func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", healthcheck)
	r.HandleFunc("/health", healthcheck)
	r.HandleFunc("/healthcheck", healthcheck)
	r.HandleFunc("/v1/webhooks/github", githubWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/webhooks/stash", stashWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/webhooks/bitbucket", bitbucketServerWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/updatePipeline", manualUpdateHandler).Methods("POST")
	return r
}

// ==============
// route handlers
// ==============

func healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Debug(r.RemoteAddr, " Requested ", r.RequestURI)
	w.Write([]byte(`{"status":"ok"}`))
}

func manualUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var fileService = dummy.FileService{}

	builder := &dinghyfile.PipelineBuilder{
		Depman:     cache.NewMemoryCache(),
		Downloader: fileService,
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	df := buf.String()
	log.Info("Received payload: ", df)
	fileService["dinghyfile"] = df

	builder.ProcessDinghyfile("", "", "dinghyfile")
}

func githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	p := github.Push{}
	if err := readBody(r, &p); err != nil {
		util.WriteHTTPError(w, err)
		return
	}

	if p.Ref == "" {
		// Unmarshal failed, might be a non-Push notification. Log event and return
		log.Debug("Possibly a non-Push notification received")
		return
	}

	buildPipelines(&p, &github.FileService{}, w)
}

func stashWebhookHandler(w http.ResponseWriter, r *http.Request) {
	payload := stash.WebhookPayload{}
	if err := readBody(r, &payload); err != nil {
		util.WriteHTTPError(w, err)
		return
	}
	log.Debugf("Payload: %+v", payload)

	payload.IsOldStash = true
	p, err := stash.NewPush(payload)
	if err != nil {
		util.WriteHTTPError(w, err)
		return
	}

	buildPipelines(p, &stash.FileService{}, w)
}

func bitbucketServerWebhookHandler(w http.ResponseWriter, r *http.Request) {
	payload := stash.WebhookPayload{}
	if err := readBody(r, &payload); err != nil {
		util.WriteHTTPError(w, err)
		return
	}

	if payload.EventKey != "" && payload.EventKey != "repo:refs_changed" {
		w.WriteHeader(200)
		return
	}

	payload.IsOldStash = false
	p, err := stash.NewPush(payload)
	if err != nil {
		util.WriteHTTPError(w, err)
		return
	}

	buildPipelines(p, &stash.FileService{}, w)
}

// =========
// utilities
// =========

// ProcessPush processes a push using a pipeline builder
func ProcessPush(p Push, b *dinghyfile.PipelineBuilder) error {
	err := error(nil)
	// Ensure dinghyfile was changed.
	if !p.ContainsFile(settings.S.DinghyFilename) {
		return nil
	}

	// Ensure we're on the master branch.
	if !p.IsMaster() {
		log.Info("Skipping Spinnaker pipeline update because this is not master")
		return nil
	}

	log.Info("Dinghyfile found in commit for repo " + p.Repo())

	// Set commit status to the pending yellow dot.
	p.SetCommitStatus(git.StatusPending)

	for _, filePath := range p.Files() {
		components := strings.Split(filePath, "/")
		if components[len(components)-1] == settings.S.DinghyFilename {
			// Process the dinghyfile.
			err = b.ProcessDinghyfile(p.Org(), p.Repo(), filePath)

			// Set commit status based on result of processing.
			if err == nil {
				p.SetCommitStatus(git.StatusSuccess)
			} else if err == dinghyfile.ErrMalformedJSON {
				p.SetCommitStatus(git.StatusFailure)
			} else {
				p.SetCommitStatus(git.StatusError)
			}
		}
	}
	return err
}

func buildPipelines(p Push, f dinghyfile.Downloader, w http.ResponseWriter) {
	// Construct a pipeline builder using provided downloader
	builder := &dinghyfile.PipelineBuilder{
		Downloader:   f,
		Depman:       GlobalCache,
		TemplateRepo: settings.S.TemplateRepo,
		TemplateOrg:  settings.S.TemplateOrg,
	}

	// Process the push.
	err := ProcessPush(p, builder)
	if err == dinghyfile.ErrMalformedJSON {
		w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		w.WriteHeader(422)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if we're in a template repo
	if p.Repo() == settings.S.TemplateRepo {
		// Set status to pending while we process modules
		p.SetCommitStatus(git.StatusPending)

		// For each module pushed, rebuild dependent dinghyfiles
		for _, file := range p.Files() {
			if err := builder.RebuildModuleRoots(p.Org(), p.Repo(), file); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				p.SetCommitStatus(git.StatusError)
				return
			}
		}
		p.SetCommitStatus(git.StatusSuccess)
	}

	w.Write([]byte(`{"status":"accepted"}`))
}

func readBody(r *http.Request, dest interface{}) error {
	body, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}
	log.Info("Received payload: ", string(body))

	util.ReadJSON(r.Body, &dest)
	return nil
}
