package statsd

import (
	"time"
)

var (
	global Client
)

// Setup enables and configures the global Statsd client.
func Setup(statsdURL string, packetSize int) (err error) {
	global, err = NewWithPacketSize(statsdURL, packetSize)
	if err != nil {
		return err
	}

	return nil
}

// Flush flushes the buffer on the global Statsd client.
func Flush() {
	if global != nil {
		global.Flush()
	}
}

// Count sends a count metric using the global Statsd client.
func Count(bucket string, value float64, sampleRate float64) {
	if global != nil {
		global.Count(bucket, value, sampleRate)
	}
}

// Gauge sends a gauge metric using the global Statsd client.
func Gauge(bucket string, value float64) {
	if global != nil {
		global.Gauge(bucket, value)
	}
}

// Timing sends a raw timing metric using the global Statsd client.
func Timing(bucket string, value float64) {
	if global != nil {
		global.Timing(bucket, value)
	}
}

// TimingDuration sends a timing metric as a time.Duration using the global
// Statsd client.
func TimingDuration(bucket string, duration time.Duration) {
	if global != nil {
		global.TimingDuration(bucket, duration)
	}
}

// CountUnique sends a unique count metric using the global Statsd client.
func CountUnique(bucket string, value string) {
	if global != nil {
		global.CountUnique(bucket, value)
	}
}
