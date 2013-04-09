/*
The statsd package provides a Statsd client. It supports all commands supported
by the Etsy statsd server implementation and automatically buffers stats into
512 byte packets.
*/
package statsd

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"regexp"
	"sync"
	"time"
)

var (
	nonAlphaNum = regexp.MustCompile(`[^\w]+`)
)

type StatsReporter interface {
	Flush() error
	Count(bucket string, value float64, sampleRate float64)
	Gauge(bucket string, value float64)
	Timing(bucket string, value time.Duration)
	CountUnique(bucket string, value string)
}

type statsdClient struct {
	PacketSize int
	prefix     string
	writer     io.Writer
	mutex      sync.Mutex
	buffer     bytes.Buffer
}

// -- emptyClient

type emptyClient struct{}

func (c emptyClient) Flush() error                   { return nil }
func (c emptyClient) Count(string, float64, float64) {}
func (c emptyClient) Gauge(string, float64)          {}
func (c emptyClient) Timing(string, time.Duration)   {}
func (c emptyClient) CountUnique(string, string)     {}

// -- statsdClient

// New connects to the given Statsd server and, optionally, uses the given
// prefix for all metric bucket names. If the prefix is "foo.bar.", a call to
// Increment with a "baz.biz" name will result in a full bucket name of
// "foo.bar.baz.biz".
//
// If there is an error resolving the host, New will return an error as well as
// a no-op StatsReporter so that code mixed with statsd calls can continue to
// run without errors.
func New(host string, prefix string) (StatsReporter, error) {
	rand.Seed(time.Now().UnixNano())
	connection, err := net.DialTimeout("udp", host, time.Second)
	if err != nil {
		return &emptyClient{}, err
	}
	return &statsdClient{
		PacketSize: 512,
		writer:     connection,
		prefix:     prefix,
	}, nil
}

func (c *statsdClient) record(sampleRate float64, bucket string, value interface{}, kind string) {
	if sampleRate < 1 && sampleRate <= rand.Float64() {
		return
	}

	if sampleRate == 1 {
		c.send(fmt.Sprintf("%s%s:%v|%s", c.prefix, bucket, value, kind))
	} else {
		c.send(fmt.Sprintf("%s%s:%v|%s|@%f", c.prefix, bucket, value, kind, sampleRate))
	}
}

func (c *statsdClient) send(data string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Flush buffer if needed
	if c.buffer.Len()+len(data)+1 >= c.PacketSize {
		err := c.Flush()
		if err != nil {
			return err
		}
	}

	// Add to buffer
	if c.buffer.Len() > 0 {
		c.buffer.WriteRune('\n')
	}
	c.buffer.WriteString(data)

	return nil
}

// Flush sends all buffered data to the statsd server, if there is any in the
// buffer.
func (c *statsdClient) Flush() error {
	if c.buffer.Len() > 0 {
		_, err := c.writer.Write(c.buffer.Bytes())
		if err != nil {
			return err
		}
		c.buffer.Reset()
	}
	return nil
}

// Gauge sets an arbitrary value. Only the value of the gauge at flush time is
// stored by statsd.
func (c *statsdClient) Gauge(bucket string, value float64) {
	c.record(1, bucket, value, "g")
}

// Count increments (or decrements) the value in a counter. Counters are
// recorded and then reset to 0 when Statsd flushes.
func (c *statsdClient) Count(bucket string, value float64, sampleRate float64) {
	c.record(sampleRate, bucket, value, "c")
}

// Timing records a time interval (in milliseconds). The percentiles, mean,
// standard deviation, sum, and lower and upper bounds are calculated by the
// Statsd server.
func (c *statsdClient) Timing(bucket string, value time.Duration) {
	c.record(1, bucket, float64(value/time.Millisecond), "ms")
}

// Unique records the number of unique values received between flushes using
// Statsd Sets.
func (c *statsdClient) CountUnique(bucket string, value string) {
	c.record(1, bucket, nonAlphaNum.ReplaceAllString(value, "_"), "s")
}
