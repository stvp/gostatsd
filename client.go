package statsd

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// Regex for sanitizing Unique() values
	nonAlpha        = regexp.MustCompile(`[^\w]+`)
	nonAlphaReplace = []byte{'_'}

	// Statsd metric type flags
	gaugeFlag       = []byte{'g'}
	countFlag       = []byte{'c'}
	timingFlag      = []byte{'m', 's'}
	cardinalityFlag = []byte{'s'}
)

// -- Client

// Client is a buffered Statsd client.
type Client interface {
	Flush() error
	Count(bucket string, value float64, sampleRate float64)
	Gauge(bucket string, value float64)
	Timing(bucket string, value float64)
	TimingDuration(bucket string, duration time.Duration)
	CountUnique(bucket string, value string)
}

// New is the same as calling NewWithPacketSize with a 512 byte packet size.
func New(statsdURL string) (Client, error) {
	return NewWithPacketSize(statsdURL, 512)
}

// NewWithPacketSize returns a new Client that will send stats to a Statsd
// server. If the given URL has a path component (eg. "/my.prefix"), all metric
// names will be prepended with that prefix.
//
// The packet size parameter is the maximum size (in bytes) that will be
// buffered before being sent. A value of 0 or less will cause each stat to be
// sent immediately, as it is received.
//
// If the Statsd URL is invalid, a no-op Client will be returned along with an
// error so that your code can ignore the error, if desired.
func NewWithPacketSize(statsdURL string, packetSize int) (Client, error) {
	// Seed random number generator for dealing with sample rates.
	rand.Seed(time.Now().UnixNano())

	host, prefix, err := parseURL(statsdURL)
	connection, err := net.DialTimeout("udp", host, time.Second)
	if err != nil {
		return &emptyClient{}, err
	}

	return &statsdClient{
		PacketSize: packetSize,
		conn:       connection,
		prefix:     []byte(prefix),
		buffer:     bytes.Buffer{},
	}, nil
}

func parseURL(statsdURL string) (host, prefix string, err error) {
	parsedStatsdURL, err := url.Parse(statsdURL)
	if err != nil {
		return "", "", err
	}
	if len(parsedStatsdURL.Host) == 0 {
		return "", "", fmt.Errorf("%#v is missing a valid hostname", statsdURL)
	}

	prefix = strings.TrimPrefix(parsedStatsdURL.Path, "/")
	if len(prefix) > 0 && prefix[len(prefix)-1] != '.' {
		prefix = prefix + "."
	}

	return parsedStatsdURL.Host, prefix, nil
}

// -- emptyClient

type emptyClient struct{}

func (c emptyClient) Flush() error                         { return nil }
func (c emptyClient) Count(string, float64, float64)       {}
func (c emptyClient) Gauge(string, float64)                {}
func (c emptyClient) Timing(string, float64)               {}
func (c emptyClient) TimingDuration(string, time.Duration) {}
func (c emptyClient) CountUnique(string, string)           {}

// -- statsdClient

type statsdClient struct {
	// Maximum size of sent UDP packets, in bytes. A value of 0 or less will
	// cause all stats to be sent immediately.
	PacketSize int

	// Prefix for all metric names. If non-blank, this should include the
	// trailing period.
	prefix []byte

	// UDP connection to Statsd
	conn net.Conn

	// Buffer metrics before sending to Statsd as UDP packets.
	buffer bytes.Buffer

	sync.Mutex
}

func (c *statsdClient) record(sampleRate float64, bucket, value, kind []byte) {
	if sampleRate < 1 && sampleRate <= rand.Float64() {
		return
	}

	sampleRateBytes := []byte{}
	if sampleRate != 1 {
		sampleRateBytes = []byte(fmt.Sprintf("|@%g", sampleRate))
	}

	if c.PacketSize <= 0 {
		c.writeMetric(bucket, value, kind, sampleRateBytes)
		c.Flush()
	} else {
		// FIXME: This is a little nasty.
		if c.buffer.Len()+1+len(c.prefix)+len(bucket)+1+len(value)+1+len(kind)+len(sampleRateBytes) > c.PacketSize {
			c.Flush()
		}
		c.writeMetric(bucket, value, kind, sampleRateBytes)
	}
}

func (c *statsdClient) writeMetric(bucket, value, kind, sampleRate []byte) {
	c.Lock()
	defer c.Unlock()

	if c.buffer.Len() > 0 {
		c.buffer.WriteRune('\n')
	}

	c.buffer.Write(c.prefix)
	c.buffer.Write(bucket)
	c.buffer.WriteRune(':')
	c.buffer.Write(value)
	c.buffer.WriteRune('|')
	c.buffer.Write(kind)
	c.buffer.Write(sampleRate)
}

// Flush sends all buffered data to the statsd server, if there is any in the
// buffer, and empties the buffer.
func (c *statsdClient) Flush() (err error) {
	c.Lock()
	defer c.Unlock()

	if c.buffer.Len() > 0 {
		_, err = c.buffer.WriteTo(c.conn)
		c.buffer.Reset()
	}
	return err
}

// Gauge sets an arbitrary value. Only the value of the gauge at flush time is
// stored by statsd.
func (c *statsdClient) Gauge(bucket string, value float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.record(1, []byte(bucket), []byte(valueString), gaugeFlag)
}

// Count increments (or decrements) the value in a counter. Counters are
// recorded and then reset to 0 when Statsd flushes.
func (c *statsdClient) Count(bucket string, value float64, sampleRate float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.record(sampleRate, []byte(bucket), []byte(valueString), countFlag)
}

// Timing records a time interval (in milliseconds). The percentiles, mean,
// standard deviation, sum, and lower and upper bounds are calculated by the
// Statsd server.
func (c *statsdClient) Timing(bucket string, value float64) {
	valueString := strconv.FormatFloat(value, 'f', -1, 64)
	c.record(1, []byte(bucket), []byte(valueString), timingFlag)
}

// TimingDuration is the same as Timing except that it takes a time.Duration
// value.
func (c *statsdClient) TimingDuration(bucket string, duration time.Duration) {
	c.Timing(bucket, float64(duration)/float64(time.Millisecond))
}

// Unique records the number of unique values received between flushes using
// Statsd Sets.
func (c *statsdClient) CountUnique(bucket string, value string) {
	cleanValue := nonAlpha.ReplaceAll([]byte(value), nonAlphaReplace)
	c.record(1, []byte(bucket), cleanValue, cardinalityFlag)
}
