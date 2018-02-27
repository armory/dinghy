package dinghyfile

import (
	"github.com/armory-io/dinghy/pkg/spinnaker"
)

// Dinghyfile is the structure of the file in an app's GitHub repo.
type Dinghyfile struct {
	Application string               `json:"application"`
	Pipelines   []spinnaker.Pipeline `json:"pipeline"`
}
