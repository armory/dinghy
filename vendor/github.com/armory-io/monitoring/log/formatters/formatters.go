package formatters

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	HTTP_FORMAT_SEPERATOR         = "--"
	HTTP_FORMAT_VALIDATION_ERRROR = "NewHttpLogFormatter called without %s field"
)

type HttpLogFormatter struct {
	Hostname   string
	CustomerID string
	Version    string
}


func NewHttpLogFormatter(hostname, customerID, version string) (*HttpLogFormatter, error) {
	fields := map[string]string{
		"hostname": hostname,
		"customerID": customerID,
		"version": version,
	}
	for attr, val := range fields {
		if val == "" {
			return nil, fmt.Errorf(HTTP_FORMAT_VALIDATION_ERRROR, attr)
		}
	}

	return &HttpLogFormatter{
		Hostname:   hostname,
		CustomerID: customerID,
		Version:    version,
	}, nil
}

func (hlf *HttpLogFormatter) Format(e *logrus.Entry) ([]byte, error) {
	parts := []string{
		hlf.Hostname,
		hlf.CustomerID,
		hlf.Version,
		"main",
		strings.ToUpper(e.Level.String()),
		"golang",
		HTTP_FORMAT_SEPERATOR,
		strings.TrimSpace(e.Message),
	}
	return []byte(strings.Join(parts, " ")), nil
}
