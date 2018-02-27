package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	log.Info("Program started.")
	serve()
}

// Serve starts the http server.
func serve() {
	http.HandleFunc("/", healthcheck)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	log.Info(r.RemoteAddr, " Requested ", r.RequestURI)
	w.Write([]byte(`{"status":"ok"}`))
}
