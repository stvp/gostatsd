package statsd

import (
	"strconv"
	"time"
)

// MockStatsdClient implements the Client interface. Instead of sending metrics
// to a Statsd, it stores received metrics for testing.
type MockStatsdClient struct {
	Counts  map[string]string
	Gauges  map[string]string
	Timings map[string]string
}

// Flush is a no-op.
func (c *MockStatsdClient) Flush() error {
	return nil
}

// Count records a Statsd count metric.
func (c *MockStatsdClient) Count(bucket string, value, sampleRate float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.Counts[bucket] = valueString
}

// Gauge records a Statsd gauge metric.
func (c *MockStatsdClient) Gauge(bucket string, value float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.Gauges[bucket] = valueString
}

// Timing records a Statsd timing metric.
func (c *MockStatsdClient) Timing(bucket string, value float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.Timings[bucket] = valueString
}

// TimingDuration records a Statsd timing metric.
func (c *MockStatsdClient) TimingDuration(bucket string, value time.Duration) {
	c.Timing(bucket, float64(value)/float64(time.Millisecond))
}

// CountUnique is currently unimplemented.
func (c *MockStatsdClient) CountUnique(bucket, value string) {
}
