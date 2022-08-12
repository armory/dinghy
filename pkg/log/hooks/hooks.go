package hooks

import (
	"bytes"
	http "github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

const HTTP_DEBUG_CONTENT_TYPE = "application/text"

type HttpDebugHook struct {
	Endpoint  string
	LogLevels []logrus.Level
	Formatter logrus.Formatter
}

func (hdk *HttpDebugHook) Levels() []logrus.Level {
	return hdk.LogLevels
}

func (hdk *HttpDebugHook) Fire(e *logrus.Entry) error {
	toSend, err := hdk.Formatter.Format(e)
	if err != nil {
		return err
	}
	go func() {
		resp, err := http.Post(hdk.Endpoint, HTTP_DEBUG_CONTENT_TYPE, bytes.NewReader(toSend))
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}()
	return nil
}
