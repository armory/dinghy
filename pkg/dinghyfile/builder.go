package dinghyfile

import (
	"encoding/json"
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
	Application          string               `json:"application"`
	DeleteStalePipelines bool                 `json:"deleteStalePipelines"`
	Pipelines            []spinnaker.Pipeline `json:"pipelines"`
}

var (
	// ErrMalformedJSON is more specific than just returning 422.
	ErrMalformedJSON = errors.New("malformed json")
)

// ProcessDinghyfile downloads a dinghyfile and uses it to update Spinnaker's pipelines.
func (b *PipelineBuilder) ProcessDinghyfile(org, repo, path string) error {

	// Render the dinghyfile and decode it into a Dinghyfile object
	buf := b.Render(org, repo, path, nil)
	log.Debug("Rendered: ", buf.String())

	d := Dinghyfile{}
	if err := json.Unmarshal(buf.Bytes(), &d); err != nil {
		log.Error("Could not unmarshal file.", err)
		return ErrMalformedJSON
	}
	log.Info("Unmarshalled: ", d)

	// TODO: validate dinghyfile

	// Update Spinnaker pipelines using received dinghyfile.
	if err := spinnaker.UpdatePipelines(d.Application, d.Pipelines, d.DeleteStalePipelines); err != nil {
		log.Error("Could not update all pipelines ", err)
		return err
	}

	return nil
}

// RebuildModuleRoots rebuilds all dinghyfiles which are roots of the specified file
func (b *PipelineBuilder) RebuildModuleRoots(org, repo, path string) error {
	url := b.Downloader.EncodeURL(org, repo, path)
	log.Info("Processing module: " + url)

	// Process all dinghyfiles that depend on this module
	for _, url := range b.Depman.GetRoots(url) {
		org, repo, path := b.Downloader.DecodeURL(url)
		if err := b.ProcessDinghyfile(org, repo, path); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}
