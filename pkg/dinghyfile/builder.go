package dinghyfile

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/spinnaker"
)

// PipelineBuilder is responsible for downloading dinghyfiles/modules, compiling them, and sending them to Spinnaker
type PipelineBuilder struct {
	Downloader   Downloader
	Depman       DependencyManager
	TemplateRepo string
	TemplateOrg  string
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

var (
	// ErrMalformedJSON is more specific than just returning 422.
	ErrMalformedJSON = errors.New("malformed json")
	DefaultEmail     = "unknown@unknown.com"
)

func UpdateDinghyfile(dinghyfile []byte) (Dinghyfile, error) {
	d := Dinghyfile{}
	if err := Unmarshal(dinghyfile, &d); err != nil {
		log.Error("Could not unmarshal dinghyfile: ", err)
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
		log.Error("Could not download Dinghyfile", err)
		return err
	}

	log.Debug("Rendered: ", buf.String())

	d, err := UpdateDinghyfile(buf.Bytes())

	log.Debug("Updated: ", buf.String())

	if err != nil {
		return err
	}

	// Update Spinnaker pipelines using received dinghyfile.
	if err := spinnaker.UpdatePipelines(d.ApplicationSpec, d.Pipelines, d.DeleteStalePipelines); err != nil {
		log.Error("Could not update all pipelines ", err)
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

	// Process all dinghyfiles that depend on this module
	for _, url := range b.Depman.GetRoots(url) {
		org, repo, path := b.Downloader.DecodeURL(url)
		if err := b.ProcessDinghyfile(org, repo, path); err != nil {
			errEncountered = true
			failedUpdates = append(failedUpdates, url)
			log.Error(err)
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
