package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"

	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/github"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/spinnaker"
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
		log.Info("Dinghyfile found in commit for repo " + p.Repo())
		p.SetCommitStatus(github.Pending)
		file, err := github.Download(settings.GitHubOrg, p.Repo(), settings.DinghyFilename)
		if err != nil {
			p.SetCommitStatus(github.Error)
			return err
		}
		log.Info("Downloaded: ", file)
		d := dinghyfile.Dinghyfile{}
		err = json.Unmarshal([]byte(file), &d)
		if err != nil {
			p.SetCommitStatus(github.Failure)
			return ErrMalformedJSON
		}
		// todo: rebuild template
		// todo: validate
		if p.IsMaster() == true {
			for _, pipeline := range d.Pipelines {
				log.Info("Updating pipeline...")
				err = spinnaker.UpdatePipeline(pipeline)
				if err != nil {
					p.SetCommitStatus(github.Error)
					return err
				}
			}
		}
		p.SetCommitStatus(github.Success)
	}
	if p.Repo() == settings.TemplateRepo {
		p.SetCommitStatus(github.Pending)
		// todo: rebuild all upstream templates.
		// todo: post them to Spinnaker
		p.SetCommitStatus(github.Success)
	}
	return nil
}
