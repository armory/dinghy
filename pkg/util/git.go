package util

import "fmt"

// GitURL returns the url from org repo and path
func GitURL(org, repo, path string) string {
	return fmt.Sprintf(`https://raw.githubusercontent.com/%s/%s/master/%s`, org, repo, path)
}
