// Package monitoring is used to monitor applications running in our internal Kubernetes clusters.
package monitoring

// Monitor used to keep track of application metrics.
type Monitor interface {
	Count(name string, value int64, tags []string, rate float64) error
	Decr(name string, tags []string, rate float64) error
	Incr(name string, tags []string, rate float64) error
	Event(title, text string) error
	Gauge(name string, value float64, tags []string, rate float64) error
}

