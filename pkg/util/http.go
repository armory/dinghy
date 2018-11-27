package util

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// WriteHTTPError writes an HTTP error to an http.ResponseWriter
func WriteHTTPError(w http.ResponseWriter, code int, err error) {
	log.Error(err)
	formattedError := fmt.Errorf(`{"status": %d, "error": "%s"}`, code, err.Error())
	http.Error(w, formattedError.Error(), code)
}
