package util

import (
	"fmt"

	"github.com/minio/minio/pkg/wildcard"
)

type GithubRateLimitErr struct {
	RateLimit int
	RateReset string
}

func (e *GithubRateLimitErr) Error() string {
	return fmt.Sprintf("Exceeded Githib rate limit: limit %d, reset %s", e.RateLimit, e.RateReset)
}

type GitHubFileNotFoundErr struct {
	Org  string
	Repo string
	Path string
}

func IsGitHubFileNotFoundErr(errString string) bool {
	pattern := "No file named * found in *"
	return wildcard.Match(pattern, errString)
}

func (e *GitHubFileNotFoundErr) Error() string {
	return fmt.Sprintf("File %s not found for org %s in repository %s", e.Path, e.Org, e.Repo)
}
