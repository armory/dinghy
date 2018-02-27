package git

// Downloader is for downloading files from a git service.
type Downloader interface {
	Download(org, repo, file string) (string, error)
}
