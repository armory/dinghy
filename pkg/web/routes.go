package web

import (
	"fmt"
	"github.com/armory-io/dinghy/pkg/modules"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"

	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/git/github"
	"github.com/armory-io/dinghy/pkg/util"
)

// Router defines the routes for the application.
func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", healthcheck)
	r.HandleFunc("/health", healthcheck)
	r.HandleFunc("/healthcheck", healthcheck)
	r.HandleFunc("/v1/webhooks/github", githubWebhookHandler).Methods("POST")
	return r
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Info(r.RemoteAddr, " Requested ", r.RequestURI)
	w.Write([]byte(`{"status":"ok"}`))
}

func githubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.DumpRequest(r, true)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"status": 500, "error": "%v"}`, err)))
		w.WriteHeader(http.StatusInternalServerError)
	}
	log.Info("Received payload: ", string(body))
	p := github.Push{}
	util.ReadJSON(r.Body, &p)
	// todo: this hangs the connection until spinnaker has been updated. shouldn't do that.
	err = dinghyfile.DownloadAndUpdate(&p, &github.FileService{})
	if err == dinghyfile.ErrMalformedJSON {
		w.Write([]byte(fmt.Sprintf(`{"error":"%v"}`, err)))
		w.WriteHeader(422)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = modules.Rebuild(&p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`{"status":"accepted"}`))
}
