// The statsd package provides a Statsd client. It supports all commands
// supported by the Etsy statsd server implementation and automatically buffers
// stats into 512 byte packets.
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
	NON_ALPHANUM         = regexp.MustCompile(`[^\w]+`)
	NON_ALPHANUM_REPLACE = []byte{'_'}

	// Statsd metric type flags
	GAUGE_FLAG       = []byte{'g'}
	COUNT_FLAG       = []byte{'c'}
	TIMING_FLAG      = []byte{'m', 's'}
	CARDINALITY_FLAG = []byte{'s'}
)

// -- Conn

type Conn struct {
	MaxPacketSize int
	Prefix        string
	network       string
	address       string
	conn          net.Conn
	buf           bytes.Buffer
	sync.Mutex
}

// NewWithPacketSize creates a new Client that will direct stats to a Statsd
// server. If the given URL has a path component (eg. "/my.prefix"), all metric
// names will be prepended with that prefix.
//
// The packet size parameter is the maximum size (in bytes) that will be
// buffered before being sent. A value of 0 or less will cause each stat to be
// sent immediately, as it is received.
func Dial(network, address string) (conn *Conn, err error) {
	// Seed random number generator for dealing with sample rates.
	rand.Seed(time.Now().UnixNano())

	conn = &Conn{
		MaxPacketSize: 512,
		buf:           bytes.Buffer{},
	}
	conn.conn, err = net.DialTimeout(network, address, time.Second)
	return conn, err
}

func (c *Conn) record(sampleRate float64, bucket, value, kind []byte) {
	if c == nil {
		return
	}

	if sampleRate < 1 && sampleRate <= rand.Float64() {
		return
	}

	sampleRateBytes := []byte{}
	if sampleRate != 1 {
		sampleRateBytes = []byte(fmt.Sprintf("|@%g", sampleRate))
	}

	if c.MaxPacketSize <= 0 {
		c.writeMetric(bucket, value, kind, sampleRateBytes)
		c.Flush()
	} else {
		// FIXME: This is a little nasty.
		if c.buf.Len()+1+len(c.Prefix)+len(bucket)+1+len(value)+1+len(kind)+len(sampleRateBytes) > c.MaxPacketSize {
			c.Flush()
		}
		c.writeMetric(bucket, value, kind, sampleRateBytes)
	}
}

func (c *Conn) writeMetric(bucket, value, kind, sampleRate []byte) {
	c.Lock()
	defer c.Unlock()

	if c.buf.Len() > 0 {
		c.buf.WriteRune('\n')
	}

	if len(c.Prefix) > 0 {
		c.buf.WriteString(c.Prefix)
	}
	c.buf.Write(bucket)
	c.buf.WriteRune(':')
	c.buf.Write(value)
	c.buf.WriteRune('|')
	c.buf.Write(kind)
	c.buf.Write(sampleRate)
}

// Flush sends all buffered data to the statsd server, if there is any in the
// buffer, and empties the buffer.
func (c *Conn) Flush() (err error) {
	c.Lock()
	defer c.Unlock()

	if c.buf.Len() > 0 {
		_, err = c.buf.WriteTo(c.conn)
		c.buf.Reset()
	}
	return err
}

// Gauge sets an arbitrary value. Only the value of the gauge at flush time is
// stored by statsd.
func (c *Conn) Gauge(bucket string, value float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.record(1, []byte(bucket), []byte(valueString), GAUGE_FLAG)
}

// Count increments (or decrements) the value in a counter. Counters are
// recorded and then reset to 0 when Statsd flushes.
func (c *Conn) Count(bucket string, value float64, sampleRate float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.record(sampleRate, []byte(bucket), []byte(valueString), COUNT_FLAG)
}

// Timing records a time interval (in milliseconds). The percentiles, mean,
// standard deviation, sum, and lower and upper bounds are calculated by the
// Statsd server.
func (c *Conn) Timing(bucket string, value float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.record(1, []byte(bucket), []byte(valueString), TIMING_FLAG)
}

// TimingDuration is the same as Timing except that it takes a time.Duration
// value.
func (c *Conn) TimingDuration(bucket string, duration time.Duration) {
	c.Timing(bucket, float64(duration)/float64(time.Millisecond))
}

// Unique records the number of unique values received between flushes using
// Statsd Sets.
func (c *Conn) CountUnique(bucket string, value string) {
	cleanValue := NON_ALPHANUM.ReplaceAll([]byte(value), NON_ALPHANUM_REPLACE)
	c.record(1, []byte(bucket), cleanValue, CARDINALITY_FLAG)
}
