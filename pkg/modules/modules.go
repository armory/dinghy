// Package modules is for the composable pieces of a dinghyfile.
package modules

import (
	"encoding/json"
	"regexp"

	"github.com/armory-io/dinghy/pkg/cache"
	"github.com/armory-io/dinghy/pkg/dinghyfile"
	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/git/github"
	"github.com/armory-io/dinghy/pkg/git/status"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/spinnaker"
	"github.com/armory-io/dinghy/pkg/util"
	log "github.com/sirupsen/logrus"
)

// Rebuild determines what modules and pipeline definitions need to be
// rebuilt based on the contents of a git push.
func Rebuild(p git.Push) error {
	c := cache.C

	if p.Repo() == settings.TemplateRepo {
		p.SetCommitStatus(status.Pending)
		files := p.Files()
		for _, f := range files {
			url := util.GitURL(p.Org(), p.Repo(), f)
			log.Info("Processing module: " + url)
			if _, exists := c[url]; exists {
				// update all upstream dinghyfiles
				_, roots := c.UpstreamNodes(c[url])
				for _, r := range roots {
					// todo: generalize and call download and update here
					err := ProcessAffectedDinghy(r.URL)
					if err != nil {
						p.SetCommitStatus(status.Error)
						return err
					}
				}
			}
		}
		p.SetCommitStatus(status.Success)
	}
	return nil
}

// ProcessAffectedDinghy downloads the affected upstream dinghyfile, renders it and
// updates the pipelines in its specification.
func ProcessAffectedDinghy(url string) error {

	r, _ := regexp.Compile("https://api.github.com/repos/(.+)/(.+)/contents/(.+)")
	match := r.FindStringSubmatch(url)
	org := match[1]
	repo := match[2]
	path := match[3]
	f := &github.FileService{}
	log.Info("Processing Dinghyfile: " + org + "/" + repo + "/" + path)
	file, err := f.Download(org, repo, path)
	if err != nil {
		log.Error("Could not download upstream dinghy file  ", err)
		return err
	}
	log.Info("Downloaded: ", file)

	// todo: handle recursive updates
	buf := dinghyfile.Render(cache.C, settings.DinghyFilename, file, org, repo, f)
	d := dinghyfile.Dinghyfile{}
	err = json.Unmarshal(buf.Bytes(), &d)
	if err != nil {
		log.Error("Could not unmarshall file: ", err)
		return err
	}
	log.Info("Unmarshalled: ", d)

	err = spinnaker.UpdatePipelines(d.Pipelines)
	if err != nil {
		log.Error("Could not update all pipelines ", err)
		return err
	}

	return nil
}
