package hooks

import (
	"bytes"
	"github.com/sirupsen/logrus"
	http "github.com/hashicorp/go-retryablehttp"
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
	if _, err := http.Post(hdk.Endpoint, HTTP_DEBUG_CONTENT_TYPE, bytes.NewReader(toSend)); err != nil {
		return err
	}
	return nil
}
