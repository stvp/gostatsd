gostatsd
========

gostatsd is a Statsd client package for Go. It supports all commands supported
by the [etsy/statsd](https://github.com/etsy/statsd/) project and automatically
buffers stats into 512 byte UDP packets.

[Documentation](http://godoc.org/github.com/stvp/gostatsd)

Usage
-----

gostatsd buffers up to 512 bytes of data before sending a UDP packet. This means
that you need to manually call `Flush()` after you're done recording your stats
to send any remaining stats.

```go
package main

import (
	"github.com/stvp/gostatsd"
	"math"
	"time"
)

func main() {
	client := statsd.New("127.0.0.1:8125", "some.prefix.")
	defer client.Flush()

	// Counters
	for i := 0; i < 10000; i++ {
		// ...
		if math.Mod(i, 100) == 0 {
			client.Count("transactions", 1, 0.100)
		}
	}

	// Gauges
	client.Gauge("queuesize", 28)

	// Timers
	start := time.Now()
	// ...
	client.Timing("methodtime", time.Since(start))
}
```

