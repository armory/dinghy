package mock

// Monitor to use in tests.
type Monitor struct {
}

// Count can be incremented or decremented.
func (m *Monitor) Count(name string, value int64, tags []string, rate float64) error {
	return nil
}

// Decr a count.
func (m *Monitor) Decr(name string, tags []string, rate float64) error {
	return nil
}

// Incr a count.
func (m *Monitor) Incr(name string, tags []string, rate float64) error {
	return nil
}

// Event marks an event.
func (m *Monitor) Event(title, text string) error {
	return nil
}

// Gauge is used to set a metric to a specific value. It will stay at that value until changed.
func (m *Monitor) Gauge(name string, value float64, tags []string, rate float64) error {
	return nil
}
