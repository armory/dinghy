// Package datadog is used to talk to datadog agents running in our Kubernetes cluster.
package datadog

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/armory-io/monitoring"
	"github.com/armory-io/monitoring/mock"
	"github.com/pborman/uuid"
	"net"
	"os"
	"strconv"
)

// Monitor is used to send metrics to a metric collection service.
type Monitor struct {
	id  string
	app string

	host   string
	port   string
	client *statsd.Client

	logger logger
	debug  bool
}

const (
	envDataDogHost    = "DATADOG_HOST"
	envDataDogPort    = "DATADOG_PORT"
	envDataDogEnabled = "DATADOG_ENABLED"
)

type logger interface {
	Printf(format string, args ...interface{})
}

// Option used to construct a new Monitor.
type Option func(*Monitor) error

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// By default return dummy monitor unless we actually want datadog enabled
func NewMonitor(opts ...Option) monitoring.Monitor {
	var m monitoring.Monitor
	m = &mock.Monitor{}
	ddEnvEnabled := getEnv(envDataDogEnabled, "false")
	ddEnabled, err := strconv.ParseBool(ddEnvEnabled)
	if err != nil {
		panic(err)
	}
	if !ddEnabled {
		return m
	}
	m = newDDClient(opts...)
	return m
}

// NewMonitor creates a monitor. It will panic if there is a problem creating it.
func newDDClient(opts ...Option) *Monitor {
	m := &Monitor{
		id:   uuid.New(),
		host: getEnv(envDataDogHost, "datadog-agent"),
		port: getEnv(envDataDogPort, "8125"),
	}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			panic(err)
		}
	}
	var err error
	m.client, err = newClient(m.host, m.port)
	if err != nil {
		panic(err)
	}
	if m.app != "" {

	}
	m.log("datadog client init")
	m.client.Tags = append(m.client.Tags, "monitor-id:"+m.id)
	if m.app != "" {
		m.client.Tags = append(m.client.Tags, "app:"+m.app)
		t := m.app + " monitor started."
		m.client.SimpleEvent(t, t)
	}
	m.log("Created datadog monitor: %+v", *m)
	return m
}

// App sets the name of the application using the monitor.
func App(name string) Option {
	return func(m *Monitor) error {
		m.app = name
		return nil
	}
}

// WithLogger sets the logger to be used by the monitor.
func WithLogger(l logger) Option {
	return func(m *Monitor) error {
		m.logger = l
		return nil
	}
}

// Debug mode for the monitor.
func Debug() Option {
	return func(m *Monitor) error {
		m.debug = true
		return nil
	}
}

func newClient(host, port string) (*statsd.Client, error) {
	c, err := statsd.New(net.JoinHostPort(host, port))
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (m *Monitor) log(format string, args ...interface{}) {
	if m.logger != nil {
		m.logger.Printf("datadog.monitor: "+format, args)
	}
}

func (m *Monitor) error(err error) {
	if m.debug == true && err != nil {
		m.log("error: %v", err)
	}
}
