package dinghyfile

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/spinnaker"
)

// PipelineBuilder is responsible for downloading dinghyfiles/modules, compiling them, and sending them to Spinnaker
type PipelineBuilder struct {
	Downloader           Downloader
	Depman               DependencyManager
	TemplateRepo         string
	TemplateOrg          string
	DinghyfileName       string
	PipelineAPI          spinnaker.PipelineAPI
	DeleteStalePipelines bool
	AutolockPipelines    string
}

// DependencyManager is an interface for assigning dependencies and looking up root nodes
type DependencyManager interface {
	SetDeps(parent string, deps []string)
	GetRoots(child string) []string
}

// Downloader is an interface that fetches files from a source
type Downloader interface {
	Download(org, repo, file string) (string, error)
	EncodeURL(org, repo, file string) string
	DecodeURL(url string) (string, string, string)
}

// Dinghyfile is the format of the pipeline template JSON
type Dinghyfile struct {
	// Application name can be specified either in top-level "application" or as a key in "spec"
	// We don't want arbitrary application properties in the top-level Dinghyfile so we put them in .spec
	Application          string                    `json:"application"`
	ApplicationSpec      spinnaker.ApplicationSpec `json:"spec"`
	DeleteStalePipelines bool                      `json:"deleteStalePipelines"`
	Globals              map[string]interface{}    `json:"globals"`
	Pipelines            []spinnaker.Pipeline      `json:"pipelines"`
}

func NewDinghyfile() Dinghyfile {
	return Dinghyfile{
		// initialize the application spec so that the default
		// enabled/disabled are initilzed slices
		// https://danott.co/posts/json-marshalling-empty-slices-to-empty-arrays-in-go.html
		ApplicationSpec: spinnaker.ApplicationSpec{
			DataSources: spinnaker.DataSourcesSpec{
				Enabled:  []string{},
				Disabled: []string{},
			},
		},
	}
}

var (
	// ErrMalformedJSON is more specific than just returning 422.
	ErrMalformedJSON = errors.New("malformed json")
	DefaultEmail     = "unknown@unknown.com"
)

func UpdateDinghyfile(dinghyfile []byte) (Dinghyfile, error) {
	d := NewDinghyfile()
	if err := Unmarshal(dinghyfile, &d); err != nil {
		return d, ErrMalformedJSON
	}
	log.Info("Unmarshalled: ", d)

	// If "spec" is not provided, these will be initialized to ""; need to pull them in.
	if d.ApplicationSpec.Name == "" {
		d.ApplicationSpec.Name = d.Application
	}
	if d.ApplicationSpec.Email == "" {
		d.ApplicationSpec.Email = DefaultEmail
	}

	return d, nil
}

// ProcessDinghyfile downloads a dinghyfile and uses it to update Spinnaker's pipelines.
func (b *PipelineBuilder) ProcessDinghyfile(org, repo, path string) error {
	// Render the dinghyfile and decode it into a Dinghyfile object
	buf, err := b.Render(org, repo, path, nil)
	if err != nil {
		return err
	}
	log.Debug("Rendered: ", buf.String())
	d, err := UpdateDinghyfile(buf.Bytes())
	if err != nil {
		return err
	}
	log.Debug("Updated: ", buf.String())

	// Update Spinnaker pipelines using received dinghyfile.
	updateOptions := spinnaker.UpdatePipelineConfig{
		DeleteStale:       d.DeleteStalePipelines,
		AutolockPipelines: b.AutolockPipelines,
	}
	if err := b.PipelineAPI.UpdatePipelines(d.ApplicationSpec, d.Pipelines, updateOptions); err != nil {
		return err
	}

	return nil
}

// RebuildModuleRoots rebuilds all dinghyfiles which are roots of the specified file
func (b *PipelineBuilder) RebuildModuleRoots(org, repo, path string) error {
	errEncountered := false
	failedUpdates := []string{}
	url := b.Downloader.EncodeURL(org, repo, path)
	log.Info("Processing module: " + url)

	// TODO: could handle logging and errors for file processing more elegantly rather
	// than making two passes.
	// Process all dinghyfiles that depend on this module
	for _, url := range b.Depman.GetRoots(url) {
		org, repo, path := b.Downloader.DecodeURL(url)
		if err := b.ProcessDinghyfile(org, repo, path); err != nil {
			errEncountered = true
			failedUpdates = append(failedUpdates, url)
		}
	}

	if errEncountered {
		log.Error("The following dinghyfiles weren't updated successfully:")
		for d := range failedUpdates {
			log.Error(d)
		}
		return errors.New("Not all upstream dinghyfiles were updated successfully")
	}
	return nil
}
