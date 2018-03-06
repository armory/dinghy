package dinghyfile

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/git/status"
	"github.com/armory-io/dinghy/pkg/settings"
	"github.com/armory-io/dinghy/pkg/spinnaker"
)

var (
	// ErrMalformedJSON is more specific than just returning 422.
	ErrMalformedJSON = errors.New("malformed json")
)

// DownloadAndUpdate downloads a dinghyfile if it is in the push notification.
// After it downloads, then it updates the pipelines in its specification.
func DownloadAndUpdate(p git.Push, f git.Downloader) error {
	if p.ContainsFile(settings.DinghyFilename) {
		log.Info("Dinghyfile found in commit for repo " + p.Repo())
		p.SetCommitStatus(status.Pending)
		file, err := f.Download(p.Org(), p.Repo(), settings.DinghyFilename)
		if err != nil {
			log.Error("Could not download dinghy file ", err)
			p.SetCommitStatus(status.Error)
			return err
		}
		log.Info("Downloaded: ", file)

		buf := Render(file, p.Org(), p.Repo(), f)

		d := Dinghyfile{}
		err = json.Unmarshal(buf.Bytes(), &d)
		if err != nil {
			log.Error("Could not unmarshall file.", err)
			p.SetCommitStatus(status.Failure)
			return ErrMalformedJSON
		}
		log.Info("Unmarshalled: ", d)
		// todo: rebuild template
		// todo: validate
		if p.IsMaster() == true {
			err = spinnaker.UpdatePipelines(d.Pipelines)
			if err != nil {
				log.Error("Could not update all pipelines ", err)
				p.SetCommitStatus(status.Error)
				return err
			}
		} else {
			log.Info("Skipping Spinnaker pipeline update because this is not master")
		}
		p.SetCommitStatus(status.Success)
	}
	return nil
}
