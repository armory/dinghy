package web

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/armory-io/dinghy/pkg/modules"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/git/github"
	"github.com/armory-io/dinghy/pkg/git/stash"
	"github.com/armory-io/dinghy/pkg/util"
)

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

func healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Info(r.RemoteAddr, " Requested ", r.RequestURI)
	w.Write([]byte(`{"status":"ok"}`))
}

func githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.DumpRequest(r, true)
	if err != nil {
		util.WriteHTTPError(w, err)
		return
	}
	log.Info("Received payload: ", string(body))
	p := github.Push{}
	util.ReadJSON(r.Body, &p)

	downloader := &github.FileService{}
	// todo: this hangs the connection until spinnaker has been updated. shouldn't do that.
	err = dinghyfile.DownloadAndUpdate(&p, downloader)
	if err == dinghyfile.ErrMalformedJSON {
		w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		w.WriteHeader(422)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = modules.Rebuild(&p, downloader); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`{"status":"accepted"}`))
}

func stashWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.DumpRequest(r, true)
	if err != nil {
		util.WriteHTTPError(w, err)
		return
	}
	log.Info("Received payload: ", string(body))

	payload := stash.WebhookPayload{}
	util.ReadJSON(r.Body, &payload)

	payload.IsOldStash = true
	p, err := stash.NewPush(payload)
	if err != nil {
		util.WriteHTTPError(w, err)
		return
	}

	downloader := &stash.FileService{}

	if err = dinghyfile.DownloadAndUpdate(p, downloader); err != nil {
		util.WriteHTTPError(w, err)
		return
	}

	if err = modules.Rebuild(p, downloader); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"status":"accepted"}`))
}

func bitbucketServerWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.DumpRequest(r, true)
	if err != nil {
		util.WriteHTTPError(w, err)
		return
	}
	log.Info("Received payload: ", string(body))

	payload := stash.WebhookPayload{}
	util.ReadJSON(r.Body, &payload)
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

	downloader := &stash.FileService{}

	if err = dinghyfile.DownloadAndUpdate(p, downloader); err != nil {
		util.WriteHTTPError(w, err)
		return
	}

	if err = modules.Rebuild(p, downloader); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"status":"accepted"}`))
}
