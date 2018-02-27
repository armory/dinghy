// Package settings is a single place to put all of the application settings.
package settings

const (
	GitHubOrg         = "armory-io"
	DinghyFilename    = "dinghyfile"
	TemplateRepo      = "dinghy-templates"
	AutoLockPipelines = true
	SpinnakerAPIURL   = "https://spinnaker.armory.io:8085"
	CertPath          = "/mnt/secrets/client.pem"

	// Temporary token. It only has access to repos and can not delete.
	GitHubUsername = "andrewbackes"
	GitHubToken    = "3ad153d626e1ffaf1bf7101d448c2b4f27d89c54"
)
