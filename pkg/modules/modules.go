// Package modules is for the composable pieces of a dinghyfile.
package modules

import (
	"github.com/armory-io/dinghy/pkg/git"
	"github.com/armory-io/dinghy/pkg/git/status"
	"github.com/armory-io/dinghy/pkg/settings"
)

// Rebuild determines what modules and pipeline definitions need to be
// rebuilt based on the contents of a git push.
func Rebuild(p git.Push) error {
	if p.Repo() == settings.TemplateRepo {
		p.SetCommitStatus(status.Pending)
		// todo: rebuild all upstream templates.
		// todo: post them to Spinnaker
		p.SetCommitStatus(status.Success)
	}
	return nil
}
