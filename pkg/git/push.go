// Package git is for working with different git services like: GitHub, GitLab, or BitBucket.
package git

import (
	"github.com/armory-io/dinghy/pkg/git/status"
)

// Push represented a push notification from a git service.
type Push interface {
	ContainsFile(file string) bool
	Repo() string
	IsMaster() bool
	SetCommitStatus(s status.Status)
}
