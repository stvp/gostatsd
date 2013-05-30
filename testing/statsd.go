package main

import (
	"time"
)

// Satisfies the StatsReporter interface to make testing easier.
type MockStatsdClient struct {
	Counts map[string]float64
	Gauges map[string]float64
}

func (c *MockStatsdClient) Flush() error {
	return nil
}

func (c *MockStatsdClient) Count(bucket string, value, sampleRate float64) {
	c.Counts[bucket] = value
}

func (c *MockStatsdClient) Gauge(bucket string, value float64) {
	c.Gauges[bucket] = value
}

func (c *MockStatsdClient) Timing(bucket string, value time.Duration) {
}

func (c *MockStatsdClient) CountUnique(bucket, value string) {
}
