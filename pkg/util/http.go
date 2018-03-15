package util

import (
	"fmt"
	"net/http"
)

// WriteHTTPError writes an HTTP error to an http.ResponseWriter
func WriteHTTPError(w http.ResponseWriter, err error) {
	w.Write([]byte(fmt.Sprintf(`{"status": 500, "error": "%v"}`, err)))
	w.WriteHeader(http.StatusInternalServerError)
}
