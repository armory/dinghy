package dummy

import "github.com/armory-io/dinghy/pkg/git/github"

// FileService serves a map[string]string of files -> file contents
type FileService map[string]string

// Download returns a file from the map
func (f FileService) Download(org, repo, file string) (string, error) {
	if ret, exists := f[file]; exists {
		return ret, nil
	}
	return "", nil
}

// EncodeURL encodes a URL
func (f FileService) EncodeURL(org, repo, path string) string {
	return (&github.FileService{}).EncodeURL(org, repo, path)
}

// DecodeURL decodes a URL
func (f FileService) DecodeURL(url string) (org, repo, path string) {
	return (&github.FileService{}).DecodeURL(url)
}
