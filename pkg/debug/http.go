package debug

import (
	"crypto/tls"
	"fmt"
	"github.com/armory/go-yaml-tools/pkg/tls/client"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

type HTTPLogger interface {
	LogRequest(string, ...interface{})
	LogResponse(string, ...interface{})
}

type InterceptorLogger struct {
	rt     http.RoundTripper
	logger HTTPLogger
}

func NewInterceptorLogger(rt http.RoundTripper, logger HTTPLogger) http.RoundTripper {
	return &InterceptorLogger{
		rt:     rt,
		logger: logger,
	}
}

func (c *InterceptorLogger) RoundTrip(req *http.Request) (*http.Response, error) {
	var u string
	u, err := url.PathUnescape(req.URL.String())
	if err != nil {
		u = req.URL.String()
	}
	c.logger.LogRequest(fmt.Sprint(fmt.Sprintf("%s --> ", req.Method), u))
	resp, err := c.rt.RoundTrip(req)
	if err != nil {
		return resp, err
	} else {
		c.logger.LogResponse(fmt.Sprint(fmt.Sprintf("%d <-- ", resp.StatusCode), u))
	}
	return resp, err
}

type LogrusDebugLogger struct {
	log *logrus.Logger
}

func NewLogrusDebugLogger(log *logrus.Logger) *LogrusDebugLogger {
	return &LogrusDebugLogger{log: log}
}

func (l *LogrusDebugLogger) LogRequest(msg string, args ...interface{}) {
	if l.log.Level == logrus.DebugLevel {
		l.log.Logf(logrus.DebugLevel, msg, args...)
	}
}

func (l *LogrusDebugLogger) LogResponse(msg string, args ...interface{}) {
	if l.log.Level == logrus.DebugLevel {
		l.log.Logf(logrus.DebugLevel, msg, args...)
	}
}

func NewInterceptorHttpClient(logger *logrus.Logger, httpOptions *client.Config, insecure bool) *http.Client {
	log := NewLogrusDebugLogger(logger)
	transport := cleanhttp.DefaultPooledTransport()
	transport.TLSClientConfig = getInterceptorTlsConfig(httpOptions, insecure)
	client := &http.Client{
		Transport: NewInterceptorLogger(transport, log),
	}
	return client
}

func getInterceptorTlsConfig(httpOptions *client.Config, insecure bool) *tls.Config {
	tlsCfg := httpOptions.GetTlsConfig()
	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}
	tlsCfg.InsecureSkipVerify = insecure
	return tlsCfg
}
