package github

type GitHubTest struct {
	contents string
	endpoint string
	err      error
}

func (g *GitHubTest) DownloadContents(org, repo, path string) (string, error) {
	return g.contents, g.err
}

func (g *GitHubTest) CreateStatus(status *Status, org, repo, ref string) error {
	return g.err
}

func (g *GitHubTest) GetEndpoint() string {
	return g.endpoint
}
