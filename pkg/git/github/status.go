package github

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/armory-io/dinghy/pkg/git/status"
	"github.com/armory-io/dinghy/pkg/settings"
)

/*
Example:

POST /repos/:owner/:repo/statuses/:sha

{
  "state": "success",
  "target_url": "https://example.com/build/status",
  "description": "The build succeeded!",
  "context": "continuous-integration/jenkins"
}

*/

type Status struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

func (p *Push) SetCommitStatus(s status.Status) {
	update := newStatus(s)
	for _, c := range p.Commits {
		sha := c.ID // not sure if this is right.
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/statuses/%s",
			p.Org(),
			p.Repo(),
			sha)
		body, err := json.Marshal(update)
		if err != nil {
			log.Debug("Could not unmarshall ", update, ": ", err)
			return
		}
		log.Info(fmt.Sprintf("Updating commit %s for %s/%s to %s.", sha, p.Org(), p.Repo(), string(s)))
		resp, err := http.Post(url, "application/json", strings.NewReader(string(body)))
		if err != nil {
			log.Error(err)
			httputil.DumpResponse(resp, true)
			return
		}
	}
}

func newStatus(s status.Status) Status {
	ret := Status{
		State:     string(s),
		TargetURL: settings.SpinnakerUIURL,
		Context:   "continuous-deployment/dinghy",
	}
	switch s {
	case status.Success:
		ret.Description = "Pipeline definitions updated!"
	case status.Error:
		ret.Description = "Error updating pipeline definitions!"
	case status.Failure:
		ret.Description = "Failed to update pipeline definitions!"
	case status.Pending:
		ret.Description = "Updating pipeline definitions..."
	}
	return ret
}
