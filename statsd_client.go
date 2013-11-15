/*
The statsd package provides a Statsd client. It supports all commands supported
by the Etsy statsd server implementation and automatically buffers stats into
512 byte packets.
*/
package statsd

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	nonAlphaNum = regexp.MustCompile(`[^\w]+`)
)

type Client interface {
	Flush() error
	Count(bucket string, value float64, sampleRate float64)
	Gauge(bucket string, value float64)
	Timing(bucket string, value time.Duration)
	CountUnique(bucket string, value string)
}

type statsdClient struct {
	// Maximum size of sent UDP packets, in bytes. A value of 0 or less will
	// cause all stats to be sent immediately.
	PacketSize int

	// Prefix for all metric names. If non-blank, this should include the
	// trailing period.
	prefix string

	// UDP connection to Statsd
	conn net.Conn

	mutex sync.Mutex

	// Buffer metrics up to a certain buffer size here.
	buffer bytes.Buffer
}

// -- emptyClient

type emptyClient struct{}

func (c emptyClient) Flush() error                   { return nil }
func (c emptyClient) Count(string, float64, float64) {}
func (c emptyClient) Gauge(string, float64)          {}
func (c emptyClient) Timing(string, time.Duration)   {}
func (c emptyClient) CountUnique(string, string)     {}

// -- Client

// New is the same as calling NewWithPacketSize with a 512 byte packet size.
func New(statsdUrl string) (Client, error) {
	return NewWithPacketSize(statsdUrl, 512)
}

// NewWithPacketSize creates a new Client that will direct stats to a Statsd
// server. If the given URL has a path component (eg. "/my.prefix"), all metric
// names will be prepended with that prefix.
//
// The packet size parameter is the maximum size (in bytes) that will be
// buffered before being sent. A value of 0 or less will cause each stat to be
// sent immediately, as it is received.
//
// If there is an error resolving the host, NewWithPacketSize will return an
// error as well as a no-op StatsReporter so that code mixed with statsd calls
// can continue to run without errors.
func NewWithPacketSize(statsdUrl string, packetSize int) (Client, error) {
	// Seed random number generator for dealing with sample rates.
	rand.Seed(time.Now().UnixNano())

	host, prefix, err := parseUrl(statsdUrl)
	connection, err := net.DialTimeout("udp", host, time.Second)
	if err != nil {
		return &emptyClient{}, err
	}

	return &statsdClient{
		PacketSize: packetSize,
		conn:       connection,
		prefix:     prefix,
	}, nil
}

func (c *statsdClient) record(sampleRate float64, bucket, value, kind string) {
	if sampleRate < 1 && sampleRate <= rand.Float64() {
		return
	}

	suffix := ""
	if sampleRate != 1 {
		suffix = fmt.Sprintf("|@%g", sampleRate)
	}

	c.send(fmt.Sprintf("%s%s:%s|%s%s", c.prefix, bucket, value, kind, suffix))
}

func (c *statsdClient) send(data string) error {
	if c.PacketSize <= 0 {
		c.writeToBuffer(data)
		c.Flush()
	} else {
		if c.buffer.Len()+len(data)+1 > c.PacketSize {
			err := c.Flush()
			if err != nil {
				return err
			}
		}
		c.writeToBuffer(data)
	}

	return nil
}

func (c *statsdClient) writeToBuffer(data string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.buffer.Len() > 0 {
		c.buffer.WriteRune('\n')
	}
	c.buffer.WriteString(data)
}

// Flush sends all buffered data to the statsd server, if there is any in the
// buffer.
func (c *statsdClient) Flush() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.buffer.Len() > 0 {
		_, err := c.conn.Write(c.buffer.Bytes())
		c.buffer.Reset()
		if err != nil {
			return err
		}
	}
	return nil
}

// Gauge sets an arbitrary value. Only the value of the gauge at flush time is
// stored by statsd.
func (c *statsdClient) Gauge(bucket string, value float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.record(1, bucket, valueString, "g")
}

// Count increments (or decrements) the value in a counter. Counters are
// recorded and then reset to 0 when Statsd flushes.
func (c *statsdClient) Count(bucket string, value float64, sampleRate float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.record(sampleRate, bucket, valueString, "c")
}

// Timing records a time interval (in milliseconds). The percentiles, mean,
// standard deviation, sum, and lower and upper bounds are calculated by the
// Statsd server.
func (c *statsdClient) Timing(bucket string, value time.Duration) {
	valueString := strconv.FormatFloat(float64(value/time.Millisecond), 'f', -1, 64)
	c.record(1, bucket, valueString, "ms")
}

// Unique records the number of unique values received between flushes using
// Statsd Sets.
func (c *statsdClient) CountUnique(bucket string, value string) {
	cleanValue := nonAlphaNum.ReplaceAllString(value, "_")
	c.record(1, bucket, cleanValue, "s")
}
