package statsd

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	MAX_PACKET_SIZE = 8932 // or 512 if you're going over the open internet
)

type StatsReporter interface {
	Count(bucket string, value int, sampleRate float32)
	Gauge(bucket string, value float32)
	Timing(bucket string, value time.Duration)
}

type statsdClient struct {
	Prefix string
	writer io.Writer
	mutex  sync.Mutex
}

// If New() can't resolve the domain name, it will return an emptyClient (and
// an error) so that all statsd functions will no-op.
type emptyClient struct{}

func (c emptyClient) Count(string, int, float32)   {}
func (c emptyClient) Gauge(string, float32)        {}
func (c emptyClient) Timing(string, time.Duration) {}

// New connects to the given Statsd server and, optionally, uses the given
// prefix for all metric bucket names. If the prefix is "foo.bar.", a call to
// Increment with a "baz.biz" name will result in a full bucket name of
// "foo.bar.baz.biz".
func New(host string, prefix string) (StatsReporter, error) {
	rand.Seed(time.Now().UnixNano())
	connection, err := net.DialTimeout("udp", host, time.Second)
	if err != nil {
		return &emptyClient{}, err
	}
	return &statsdClient{writer: connection, Prefix: prefix}, nil
}

func (c *statsdClient) record(sampleRate float32, bucket string, delta float32, kind string) {
	if sampleRate < 1 && sampleRate <= rand.Float32() {
		return
	}
	if sampleRate != 1 {
		c.send(fmt.Sprintf("%s%s:%v|%s|@%f", c.Prefix, bucket, delta, kind, sampleRate))
	} else {
		c.send(fmt.Sprintf("%s%s:%v|%s", c.Prefix, bucket, delta, kind))
	}
}

func (c *statsdClient) send(data string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, err := c.writer.Write([]byte(data))
	if err != nil {
		return err
	}
	return nil
}

// Gauge sets an arbitrary value.
func (c *statsdClient) Gauge(bucket string, value float32) {
	c.record(1, bucket, value, "g")
}

// Count increments (or decrements the value in a counter). Counters are
// recorded and then reset to 0 when Statsd flushes.
func (c *statsdClient) Count(bucket string, value int, sampleRate float32) {
	c.record(sampleRate, bucket, float32(value), "c")
}

// Timing records a time interval (in milliseconds). The
// percentiles, mean, standard deviation, sum, and lower and upper
// bounds are calculated by the Statsd server.
func (c *statsdClient) Timing(bucket string, value time.Duration) {
	c.record(1, bucket, float32(value/time.Millisecond), "ms")
}
