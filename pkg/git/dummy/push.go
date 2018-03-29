package dummy

import "github.com/armory-io/dinghy/pkg/git"

// Push contains data about a push full of commits
type Push struct {
	RepoName  string
	OrgName   string
	FileNames []string
}

// ContainsFile checks to see if a given file is in the push.
func (p *Push) ContainsFile(file string) bool {
	for _, name := range p.FileNames {
		if name == file {
			return true
		}
	}
	return false
}

// Files returns a slice containing filenames that were added/modified
func (p *Push) Files() []string {
	return p.FileNames
}

// Repo returns the name of the repo.
func (p *Push) Repo() string {
	return p.RepoName
}

// Org returns the name of the project.
func (p *Push) Org() string {
	return p.OrgName
}

// IsMaster detects if the branch is master.
func (p *Push) IsMaster() bool {
	return true
}

// SetCommitStatus sets a commit status
func (p *Push) SetCommitStatus(s git.Status) {
	// Do nothing
}
