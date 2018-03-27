// Package modules is for the composable pieces of a dinghyfile.
package modules

import (
	"encoding/json"
	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/git/status"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/spinnaker"
	log "github.com/sirupsen/logrus"
)

// Rebuild determines what modules and pipeline definitions need to be
// rebuilt based on the contents of a git push.
func Rebuild(p git.Push, downloader git.Downloader) error {
	c := cache.C

	if p.Repo() == settings.S.TemplateRepo {
		p.SetCommitStatus(status.Pending)
		files := p.Files()
		for _, f := range files {
			url := downloader.GitURL(p.Org(), p.Repo(), f)
			log.Info("Processing module: " + url)

			_, rootURLs := c.UpstreamURLs(url)
			for _, r := range rootURLs {
				// todo: generalize and call download and update here
				err := ProcessAffectedDinghy(r, downloader)
				if err != nil {
					p.SetCommitStatus(status.Error)
					return err
				}
			}
		}
		p.SetCommitStatus(status.Success)
	}
	return nil
}

// ProcessAffectedDinghy downloads the affected upstream dinghyfile, renders it and
// updates the pipelines in its specification.
func ProcessAffectedDinghy(url string, downloader git.Downloader) error {
	org, repo, path := downloader.ParseGitURL(url)
	log.Info("Processing Dinghyfile: " + org + "/" + repo + "/" + path)
	file, err := downloader.Download(org, repo, path)
	if err != nil {
		log.Error("Could not download upstream dinghy file  ", err)
		return err
	}
	log.Info("Downloaded: ", file)

	// todo: handle recursive updates
	buf := dinghyfile.Render(cache.C, settings.S.DinghyFilename, file, org, repo, downloader)
	d := dinghyfile.Dinghyfile{}
	err = json.Unmarshal(buf.Bytes(), &d)
	if err != nil {
		log.Error("Could not unmarshall file: ", err)
		return err
	}
	log.Info("Unmarshalled: ", d)

	err = spinnaker.UpdatePipelines(d.Application, d.Pipelines)
	if err != nil {
		log.Error("Could not update all pipelines ", err)
		return err
	}

	return nil
}
