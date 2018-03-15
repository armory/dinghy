package stash

import (
	"github.com/armory-io/dinghy/pkg/git/status"
)

// SetCommitStatus sets a commit status
func (p *Push) SetCommitStatus(s status.Status) {
	// todo, does stash even have commit statuses?
}
