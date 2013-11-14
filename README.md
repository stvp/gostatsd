gostatsd
========

gostatsd is a Statsd client package for Go. It supports all commands supported
by the [etsy/statsd](https://github.com/etsy/statsd/) project and automatically
buffers stats into 512 byte UDP packets. You can customize the packet size, as
well.

[Documentation](http://godoc.org/github.com/stvp/gostatsd)

Usage
-----

gostatsd, by default, buffers up to 512 bytes of data before sending a UDP
packet. This means that you need to manually call `Flush()` after you're done
recording your stats to send any remaining stats.

```go
package main

import (
	"github.com/stvp/gostatsd"
	"math"
	"time"
)

func main() {
	statsd.Setup("statsd://127.0.0.1:8125/my.prefix", 512)
	defer statsd.Flush()

	// Counters
	for i := 0; i < 10000; i++ {
		if math.Mod(i, 100) == 0 {
			statsd.Count("transactions", 1, 0.100)
		}
	}

	// Gauges
	statsd.Gauge("queuesize", 28)

	// Timers
	start := time.Now()
	statsd.Timing("methodtime", time.Since(start))
}
```

You can also create a Client object and use that, instead:

```go
package main

import (
	"github.com/stvp/gostatsd"
	"math"
	"time"
)

func main() {
	client := statsd.New("statsd://127.0.0.1:8125", "some.prefix.")
	defer client.Flush()

	// Counters
	for i := 0; i < 10000; i++ {
		if math.Mod(i, 100) == 0 {
			client.Count("transactions", 1, 0.100)
		}
	}

	// Gauges
	client.Gauge("queuesize", 28)

	// Timers
	start := time.Now()
	client.Timing("methodtime", time.Since(start))
}
```

The buffer size (in bytes) can be customized:

```go
client := NewWithPacketSize("127.0.0.1:8125", "my.prefix.", 128)
```

### Unbuffered sending

A buffer size of 0 or less will cause all stats to be sent individually as soon
as they are received. (This also means that you wont need to call `Flush`.)

```go
statsd.Setup("127.0.0.1:8125/my.prefix", -1)
// or
client := statsd.NewWithPacketSize("127.0.0.1:8125", "my.prefix.", -1)
```

