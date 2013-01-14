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
	Increment(name string, sampleRate float32)
	Decrement(name string, sampleRate float32)
	Count(name string, value int, sampleRate float32)
	Gauge(name string, value float32)
	Timing(name string, value time.Duration)
}

type statsdClient struct {
	Prefix string
	writer io.Writer
	mutex  sync.Mutex
}

type emptyClient struct{}

func (c emptyClient) Increment(string, float32)    {}
func (c emptyClient) Decrement(string, float32)    {}
func (c emptyClient) Count(string, int, float32)   {}
func (c emptyClient) Gauge(string, float32)        {}
func (c emptyClient) Timing(string, time.Duration) {}

// New connects to the given Statsd server and, optionally, uses the given
// namespace as a prefix for all supplied metric names. For example, if the
// namespace is "foo.bar", a call to Increment with a "baz.biz" name will
// result in a full path of "foo.bar.baz.biz".
func New(host string, namespace string) (StatsReporter, error) {
	connection, err := net.DialTimeout("udp", host, time.Second)
	if err != nil {
		return &emptyClient{}, err
	}
	if len(namespace) > 0 {
		return &statsdClient{writer: connection, Prefix: namespace + "."}, nil
	}
	return &statsdClient{writer: connection, Prefix: ""}, nil
}

func (c *statsdClient) record(sampleRate float32, name string, delta float32, kind string) {
	// TODO seed rand?
	// TODO scrub name
	if sampleRate < 1 && sampleRate <= rand.Float32() {
		return
	}
	if sampleRate != 1 {
		c.send(fmt.Sprintf("%s%s:%v|%s|@%f", c.Prefix, name, delta, kind, sampleRate))
	} else {
		c.send(fmt.Sprintf("%s%s:%v|%s", c.Prefix, name, delta, kind))
	}
	// TODO buffer and flush
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

func (c *statsdClient) Gauge(name string, value float32) {
	c.record(1, name, value, "g")
}

func (c *statsdClient) Count(name string, value int, sampleRate float32) {
	c.record(sampleRate, name, float32(value), "c")
}

func (c *statsdClient) Increment(name string, sampleRate float32) {
	c.record(sampleRate, name, 1, "c")
}

func (c *statsdClient) Decrement(name string, sampleRate float32) {
	c.record(sampleRate, name, -1, "c")
}

func (c *statsdClient) Timing(name string, value time.Duration) {
	c.record(1, name, float32(value/time.Millisecond), "ms")
}

// https://github.com/etsy/statsd
// TODO: Sets
// TODO: and so on
