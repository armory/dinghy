package git

// Downloader is for downloading files from a git service.
type Downloader interface {
	Download(org, repo, file string) (string, error)
	GitURL(org, repo, file string) string
	ParseGitURL(url string) (string, string, string)
}
