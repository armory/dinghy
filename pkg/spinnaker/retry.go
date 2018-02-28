package spinnaker

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

func postWithRetry(url string, body []byte) (resp *http.Response, err error) {
	for retry := 0; retry < 10; retry++ {
		log.Debug("POST ", url)
		resp, err = defaultClient.Post(url, "application/json", strings.NewReader(string(body)))
		timeout := time.Duration(retry*200) * time.Millisecond
		if err != nil {
			log.Error(err)
			time.Sleep(timeout)
			continue
		}
		if resp.StatusCode > 399 {
			log.Error("Spinnaker returned ", resp.StatusCode)
			httputil.DumpResponse(resp, true)
			time.Sleep(timeout)
			continue
		}
		break
	}
	if resp != nil && resp.StatusCode > 399 {
		err = fmt.Errorf("spinnaker returned %d", resp.StatusCode)
	}
	return resp, nil
}

func getWithRetry(url string) (resp *http.Response, err error) {
	for retry := 0; retry < 10; retry++ {
		log.Debug("GET ", url)
		resp, err = defaultClient.Get(url)
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
			httputil.DumpResponse(resp, true)
			time.Sleep(timeout)
			continue
		}
		break
	}
	if resp != nil && resp.StatusCode > 399 {
		err = fmt.Errorf("spinnaker returned %d", resp.StatusCode)
	}
	return
}
