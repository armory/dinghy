package web

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"

	"github.com/armory-io/dinghy/pkg/github"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/util"
)

var (
	// ErrMalformedJSON is more specific than just returning 422.
	ErrMalformedJSON = errors.New("malformed json")
)

// Router defines the routes for the application.
func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", healthcheck)
	r.HandleFunc("/health", healthcheck)
	r.HandleFunc("/healthcheck", healthcheck)
	r.HandleFunc("/v1/webhooks/github", webhookHandler).Methods("POST")
	return r
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Info(r.RemoteAddr, " Requested ", r.RequestURI)
	w.Write([]byte(`{"status":"ok"}`))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.DumpRequest(r, true)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"status": 500, "error": "%v"}`, err)))
		w.WriteHeader(http.StatusInternalServerError)
	}
	log.Info("Received payload: ", string(body))
	p := github.Payload{}
	util.ReadJSON(r.Body, &p)
	err = processPayload(p)
	if err == ErrMalformedJSON {
		w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		w.WriteHeader(422)
		return
	}
	w.Write([]byte(`{"status":"accepted"}`))
}

func processPayload(p github.Payload) error {
	if p.ContainsFile(settings.DinghyFilename) {
		// todo: update commit message to in progress
		file, err := github.Download(settings.GitHubOrg, p.Repo(), settings.DinghyFilename)
		if err != nil {
			return ErrMalformedJSON
		}
		log.Info("Downloaded: ", file)
		// todo: rebuild template
		// todo: validate
		// todo: update commit message to green check or red x
		if p.IsMaster() == true {
			// todo: post to Spinnaker
		}
	}
	if p.Repo() == settings.TemplateRepo {
		// todo: rebuild all upstream templates.
		// todo: post them to Spinnaker
	}
	return nil
}
