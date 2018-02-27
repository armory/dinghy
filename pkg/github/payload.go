package github

// Payload is received from a GitHub webhook.
type Payload struct {
	Commits []Commit `json:"commits"`
}

type Commit struct {
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
}

// Contains checks to see if a given file is in the payload.
func (p *Payload) Contains(file string) bool {
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

// IsRepo detects if the payload has to do with a certain repo.
func (p *Payload) IsRepo(repo string) bool {
	return false
}
