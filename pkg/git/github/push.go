package github

// Push is the payload received from a GitHub webhook.
type Push struct {
	Commits    []Commit   `json:"commits"`
	Repository Repository `json:"repository"`
	Ref        string     `json:"ref"`
}

// Commit is a commit received from Github webhook
type Commit struct {
	ID       string   `json:"id"`
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
}

// Repository is a repo received from Github webhook
type Repository struct {
	Name         string `json:"name"`
	Organization string `json:"organization"`
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

// Files returns a slice containing filenames that were added/modified
func (p *Push) Files() []string {
	ret := make([]string, 0, 0)
	if p.Commits == nil {
		return ret
	}
	for _, c := range p.Commits {
		ret = append(ret, c.Added...)
		ret = append(ret, c.Modified...)
	}
	return ret
}

// Repo returns the name of the repo.
func (p *Push) Repo() string {
	return p.Repository.Name
}

// Org returns the organization of the push
func (p *Push) Org() string {
	return p.Repository.Organization
}

// IsMaster detects if the branch is master.
func (p *Push) IsMaster() bool {
	return p.Ref == "refs/heads/master"
}
