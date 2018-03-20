package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// WriteHTTPError writes an HTTP error to an http.ResponseWriter
func WriteHTTPError(w http.ResponseWriter, err error) {
	log.Error(err)

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(`{"status": 500, "error": "%v"}`, err)))
}
