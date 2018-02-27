package github

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	// Temporary token. It only has access to repos and can not delete.
	username = "andrewbackes"
	token    = "3ad153d626e1ffaf1bf7101d448c2b4f27d89c54"
)

// Download a file from github.
func Download(org, repo, file string) (string, error) {
	url := fmt.Sprintf(`https://raw.githubusercontent.com/%s/%s/master/%s`, org, repo, file)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "token "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
