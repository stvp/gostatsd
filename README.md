GoStatsd
========

[![Build Status](https://travis-ci.org/stvp/gostatsd.png)](https://travis-ci.org/stvp/gostatsd)

GoStatsd is a simple Statsd client package for Go.

Usage
-----

```go
import "github.com/stvp/gostatsd"
client := statsd.New("127.0.0.1:8125", "some.prefix.")

client.Count("transactions", 15, 1)
client.Gauge("queuesize", 28)

start := time.Now()
// ...
client.Timing("methodtime", time.Since(start))
```

TODO
----

* Support buffering and sending as 512 byte packets
* Support "extra" stats methods provided by different Statsd servers like
  Sets

