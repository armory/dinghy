package web

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/git/github"
	"github.com/armory-io/dinghy/pkg/git/stash"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/util"
)

// Push represented a push notification from a git service.
type Push interface {
	ContainsFile(file string) bool
	Files() []string
	Repo() string
	Org() string
	IsMaster() bool
	SetCommitStatus(s git.Status)
}

// Router defines the routes for the application.
func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", healthcheck)
	r.HandleFunc("/health", healthcheck)
	r.HandleFunc("/healthcheck", healthcheck)
	r.HandleFunc("/v1/webhooks/github", githubWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/webhooks/stash", stashWebhookHandler).Methods("POST")
	r.HandleFunc("/v1/webhooks/bitbucket", bitbucketServerWebhookHandler).Methods("POST")
	return r
}

// ==============
// route handlers
// ==============

func healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Info(r.RemoteAddr, " Requested ", r.RequestURI)
	w.Write([]byte(`{"status":"ok"}`))
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

	if payload.EventKey != nil && *payload.EventKey != "repo:refs_changed" {
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

func processPush(p Push, b *dinghyfile.PipelineBuilder) error {
	// Ensure dinghyfile was changed.
	if !p.ContainsFile(settings.S.DinghyFilename) {
		p.SetCommitStatus(git.StatusSuccess)
		return nil
	}

	// Ensure we're on the master branch.
	if !p.IsMaster() {
		log.Info("Skipping Spinnaker pipeline update because this is not master")
		p.SetCommitStatus(git.StatusSuccess)
		return nil
	}

	// Set commit status to the pending yellow dot.
	p.SetCommitStatus(git.StatusPending)

	// Process the dinghyfile.
	err := b.ProcessDinghyfile(p.Org(), p.Repo(), settings.S.DinghyFilename)

	// Set commit status based on result of processing.
	if err == nil {
		p.SetCommitStatus(git.StatusSuccess)
	} else if err == dinghyfile.ErrMalformedJSON {
		p.SetCommitStatus(git.StatusFailure)
	} else {
		p.SetCommitStatus(git.StatusError)
	}

	return err
}

func buildPipelines(p Push, f dinghyfile.Downloader, w http.ResponseWriter) {
	// Construct a pipeline builder using provided downloader
	builder := dinghyfile.NewPipelineBuilder(f, cache.C)

	// Process the push.
	err := processPush(p, builder)
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
