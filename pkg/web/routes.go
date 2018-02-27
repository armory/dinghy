package web

import (
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"

	"github.com/armory-io/dinghy/pkg/github"
	"github.com/armory-io/dinghy/pkg/util"
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
	processPayload(p)
	w.Write([]byte(`{"status":"accepted"}`))
}

const (
	dinghyFile   = "dinghyfile"
	templateRepo = "dinghy-templates"
)

func processPayload(p github.Payload) {
	if p.Contains(dinghyFile) {
		// update commit message to in progress
		// rebuild template
		// validate
		// update commit message to green check or red x
		if p.IsMaster() {
			// post to Spinnaker
		}
	}
	if p.IsRepo(templateRepo) {
		// rebuild all upstream templates.
		// post them to Spinnaker
	}
}
