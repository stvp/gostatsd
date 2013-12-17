package statsd

import (
	"time"
)

var (
	client Client
)

func Setup(statsdUrl string, packetSize int) (err error) {
	client, err = NewWithPacketSize(statsdUrl, packetSize)
	if err != nil {
		return err
	}

	return nil
}

func Flush() {
	if client != nil {
		client.Flush()
	}
}

func Count(bucket string, value float64, sampleRate float64) {
	if client != nil {
		client.Flush()
	}
}

func Gauge(bucket string, value float64) {
	if client != nil {
		client.Gauge(bucket, value)
	}
}

func Timing(bucket string, value float64) {
	if client != nil {
		client.Timing(bucket, value)
	}
}

func TimingDuration(bucket string, duration time.Duration) {
	if client != nil {
		client.TimingDuration(bucket, duration)
	}
}

func CountUnique(bucket string, value string) {
	if client != nil {
		client.CountUnique(bucket, value)
	}
}
