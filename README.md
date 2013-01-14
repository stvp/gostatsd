GoStatsd
========

[![Build Status](https://travis-ci.org/stvp/gostatsd.png)](https://travis-ci.org/stvp/gostatsd)

GoStatsd is a simple Statsd client package for Go. It supports all commands
supported by the [etsy/statsd](https://github.com/etsy/statsd/) project.

Usage
-----

```go
client := statsd.New("127.0.0.1:8125", "some.prefix.")

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
```

TODO
----

* Support buffering and sending as 512 byte packets

