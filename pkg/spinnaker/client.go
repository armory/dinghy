package spinnaker

import (
	"net/http"
)

var defaultClient *http.Client

func init() {
	defaultClient = &http.Client{}
}
