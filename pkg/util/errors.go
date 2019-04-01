package util

import (
	"fmt"
)

type GithubRateLimitErr struct {
	RateLimit int
	RateReset string
}

func (e *GithubRateLimitErr) Error() string {
	return fmt.Sprintf("Exceeded Githib rate limit: limit %d, reset %s", e.RateLimit, e.RateReset)
}
