package util

import "fmt"

// GitURL returns the url from org repo and path
func GitURL(org, repo, path string) string {
	return fmt.Sprintf(`https://api.github.com/repos/%s/%s/contents/%s`, org, repo, path)
}
