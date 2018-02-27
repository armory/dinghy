package github

import (
	"github.com/armory-io/dinghy/pkg/git/status"
)

// Payload is received from a GitHub webhook.
type Push struct {
	Commits    []Commit   `json:"commits"`
	Repository Repository `json:"repository"`
	Ref        string     `json:"ref"`
}

type Commit struct {
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
}

type Repository struct {
	Name string `json:"name"`
}

// ContainsFile checks to see if a given file is in the push.
func (p *Push) ContainsFile(file string) bool {
	if p.Commits == nil {
		return false
	}
	for _, c := range p.Commits {
		for _, f := range c.Added {
			if f == file {
				return true
			}
		}
		for _, f := range c.Modified {
			if f == file {
				return true
			}
		}
	}
	return false
}

// Repo returns the name of the repo.
func (p *Push) Repo() string {
	return p.Repository.Name
}

// IsMaster detects if the branch is master.
func (p *Push) IsMaster() bool {
	return p.Ref == "refs/heads/master"
}

func (p *Push) SetCommitStatus(s status.Status) error {
	// todo: for each commit
	return nil
}
