package spinnaker

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/armory-io/dinghy/pkg/settings"
	log "github.com/sirupsen/logrus"
)

type callback func() (*http.Response, error)

func requestWithRetry(cb callback) (resp *http.Response, err error) {
	for retry := 0; retry < 10; retry++ {
		resp, err = cb()
		timeout := time.Duration(retry*200) * time.Millisecond
		if err != nil {
			log.Error(err)
			if resp != nil {
				httputil.DumpResponse(resp, true)
			}
			time.Sleep(timeout)
			continue
		}
		if resp.StatusCode > 399 {
			log.Error("Spinnaker returned ", resp.StatusCode)
			body, _ := ioutil.ReadAll(resp.Body)
			log.Print(string(body))
			httputil.DumpResponse(resp, true)
			time.Sleep(timeout)
			continue
		}
		break
	}
	if resp != nil && resp.StatusCode > 399 {
		err = fmt.Errorf("spinnaker returned %d", resp.StatusCode)
	}
	return resp, err
}

func postWithRetry(url string, body []byte) (resp *http.Response, err error) {
	return requestWithRetry(func() (*http.Response, error) {
		log.Debug("POST ", url)
		return request("POST", url, strings.NewReader(string(body)))
	})
}

func getWithRetry(url string) (resp *http.Response, err error) {
	return requestWithRetry(func() (*http.Response, error) {
		log.Debug("GET ", url)
		return request("GET", url, nil)
	})
}

func request(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/context+json")
	if settings.S.Fiat.Enabled && settings.S.Fiat.AuthUser != "" {
		req.Header.Set("X-Spinnaker-User", settings.S.Fiat.AuthUser)
	}
	return defaultClient.Do(req)
}

func deleteWithRetry(url string) (resp *http.Response, err error) {
	return requestWithRetry(func() (*http.Response, error) {
		log.Debug("DELETE ", url)
		return request("DELETE", url, nil)
	})
}

func putWithRetry(url string, body []byte) (resp *http.Response, err error) {
	return requestWithRetry(func() (*http.Response, error) {
		log.Debug("PUT ", url)
		return request("PUT", url, strings.NewReader(string(body)))
	})
}
