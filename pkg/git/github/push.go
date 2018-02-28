package github

// Payload is received from a GitHub webhook.
type Push struct {
	Commits    []Commit   `json:"commits"`
	Repository Repository `json:"repository"`
	Ref        string     `json:"ref"`
}

type Commit struct {
	ID       string   `json:"id"`
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
}

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

// Repo returns the name of the repo.
func (p *Push) Repo() string {
	return p.Repository.Name
}

func (p *Push) Org() string {
	return p.Repository.Organization
}

// IsMaster detects if the branch is master.
func (p *Push) IsMaster() bool {
	return p.Ref == "refs/heads/master"
}
