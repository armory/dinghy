package datadog

// Count can be incremented or decremented.
func (m *Monitor) Count(name string, value int64, tags []string, rate float64) error {
	m.log("Count{name: %s, value: %d, tags: %v, rate: %v}", name, value, tags, rate)
	err := m.client.Count(name, value, tags, rate)
	m.error(err)
	return err
}

// Decr a count.
func (m *Monitor) Decr(name string, tags []string, rate float64) error {
	m.log("DecrCount{name: %s, tags: %v, rate: %v}", name, tags, rate)
	err := m.client.Decr(name, tags, rate)
	m.error(err)
	return err
}

// Incr a count.
func (m *Monitor) Incr(name string, tags []string, rate float64) error {
	m.log("IncrCount{name: %s, tags: %v, rate: %v}", name, tags, rate)
	err := m.client.Incr(name, tags, rate)
	m.error(err)
	return err
}

// Event marks an event.
func (m *Monitor) Event(title, text string) error {
	m.log("Event{title: %s, text: %s}", title, text)
	err := m.client.SimpleEvent(title, text)
	m.error(err)
	return err
}

// Gauge is used to set a metric to a specific value. It will stay at that value until changed.
func (m *Monitor) Gauge(name string, value float64, tags []string, rate float64) error {
	m.log("Gauge{name: %s, value: %d, tags: %v, rate: %v}", name, value, tags, rate)
	err := m.client.Gauge(name, value, tags, rate)
	m.error(err)
	return err
}
