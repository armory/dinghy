package github

// Payload is received from a GitHub webhook.
type Payload struct {
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

// Status wires up to the green check or red x next to a GitHub commit.
type Status string

const (
	Pending Status = "pending"
	Error          = "error"
	Success        = "success"
	Failure        = "failure"
)

// ContainsFile checks to see if a given file is in the payload.
func (p *Payload) ContainsFile(file string) bool {
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
func (p *Payload) Repo() string {
	return p.Repository.Name
}

// IsMaster detects if the branch is master.
func (p *Payload) IsMaster() bool {
	return p.Ref == "refs/heads/master"
}

func (p *Payload) SetCommitStatus(s Status) error {
	// todo: for each commit
	return nil
}
