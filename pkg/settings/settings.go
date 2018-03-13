// Package settings is a single place to put all of the application settings.
package settings

import (
	"os"

	"github.com/armory-io/dinghy/pkg/util"
)

var (
	GitHubOrg         = "armory-io"
	DinghyFilename    = "dinghyfile"
	TemplateRepo      = "dinghy-templates"
	AutoLockPipelines = true
	SpinnakerAPIURL   = "https://spinnaker.armory.io:8085"
	SpinnakerUIURL    = "https://spinnaker.armory.io"
	CertPath          = util.GetenvOrDefault("CLIENT_CERT_PATH", os.Getenv("HOME")+"/.armory/cache/client.pem")

	// Temporary token. It only has access to repos and can not delete.
	GitHubUsername = "andrewbackes"
	GitHubToken    = "3ad153d626e1ffaf1bf7101d448c2b4f27d89c54"
)
