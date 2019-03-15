package spinnaker

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type APIClient interface {
	RequestWithRetry(cb callback) (*http.Response, error)
	PostWithRetry(url string, body []byte) (*http.Response, error)
	GetWithRetry(url string) (*http.Response, error)
	DeleteWithRetry(url string) (*http.Response, error)
	Request(method, url string, body io.Reader) (*http.Response, error)
}

type DefaultAPIClient struct {
	Headers map[string]string
}

type callback func() (*http.Response, error)

func (client *DefaultAPIClient) RequestWithRetry(cb callback) (resp *http.Response, err error) {
	for retry := 0; retry < 10; retry++ {
		resp, err = cb()
		timeout := time.Duration(retry*200) * time.Millisecond
		if err != nil {
			if resp != nil {
				// TODO: we shouldn't be using DumpResponse here
				httputil.DumpResponse(resp, true)
			}
			time.Sleep(timeout)
			continue
		}
		if resp.StatusCode > 399 {
			// TODO: we shouldnt be using DumpResponse here
			httputil.DumpResponse(resp, true)
			time.Sleep(timeout)
			continue
		}
		break
	}
	if resp != nil && resp.StatusCode > 399 {
		return nil, fmt.Errorf("spinnaker returned %d: %s", resp.StatusCode, resp.Body)
	}
	return resp, err
}

func (client *DefaultAPIClient) PostWithRetry(url string, body []byte) (resp *http.Response, err error) {
	return client.RequestWithRetry(func() (*http.Response, error) {
		log.Info("POST ", url)
		log.Debug("BODY ", string(body))
		return client.Request("POST", url, strings.NewReader(string(body)))
	})
}

func (client *DefaultAPIClient) GetWithRetry(url string) (resp *http.Response, err error) {
	return client.RequestWithRetry(func() (*http.Response, error) {
		log.Info("GET ", url)
		return client.Request("GET", url, nil)
	})
}

func (client *DefaultAPIClient) Request(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/context+json")
	for key, value := range client.Headers {
		req.Header.Set(key, value)
	}

	return defaultClient.Do(req)
}

func (client *DefaultAPIClient) DeleteWithRetry(url string) (resp *http.Response, err error) {
	return client.RequestWithRetry(func() (*http.Response, error) {
		log.Info("DELETE ", url)
		return client.Request("DELETE", url, nil)
	})
}
