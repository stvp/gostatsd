package statsd

import (
	"time"
)

var (
	conn *Conn
)

func Setup(network, address string) (err error) {
	conn, err = Dial(network, address)
	return err
}

func Flush() {
	if conn != nil {
		conn.Flush()
	}
}

func Count(bucket string, value float64, sampleRate float64) {
	if conn != nil {
		conn.Count(bucket, value, sampleRate)
	}
}

func Gauge(bucket string, value float64) {
	if conn != nil {
		conn.Gauge(bucket, value)
	}
}

func Timing(bucket string, value float64) {
	if conn != nil {
		conn.Timing(bucket, value)
	}
}

func TimingDuration(bucket string, duration time.Duration) {
	if conn != nil {
		conn.TimingDuration(bucket, duration)
	}
}

func CountUnique(bucket string, value string) {
	if conn != nil {
		conn.CountUnique(bucket, value)
	}
}
