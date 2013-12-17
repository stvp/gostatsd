gostatsd
========

gostatsd is a Statsd client package for Go. It supports all commands supported
by the [etsy/statsd](https://github.com/etsy/statsd/) project and automatically
buffers stats into 512 byte UDP packets. You can customize the packet size, as
well.

[API documentation](http://godoc.org/github.com/stvp/gostatsd)

Usage
-----

gostatsd, by default, buffers up to 512 bytes of data before sending a UDP
packet. This means that you need to manually call `Flush()` after you're done
recording your stats to send any remaining stats.

Basic usage is easy: just give gostatsd a URL and a packet size, and start
sending metrics:

```go
statsd.Setup("statsd://127.0.0.1:8125/prefix.here", 512)
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
statsd.Timing("methodtime", float64(time.Since(start))/float64(time.Milliseconds))
// or
statsd.TimingDuration("methodtime", time.Since(start))
```

If you need to send metrics to different places or want to use different metric
prefixes, you can create a standalone Client:

```go
client := statsd.New("statsd://127.0.0.1:8125/my.prefix")
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
client.TimingDuration("methodtime", time.Since(start))
```

The buffer size (in bytes) can be customized:

```go
client := NewWithPacketSize("statsd://127.0.0.1:8125/my.prefix.", 128)
```

### Unbuffered sending

A buffer size of 0 or less will cause all stats to be sent individually as soon
as they are received. If you're using unbuffered sending, you wont need to call
`Flush()`

```go
statsd.Setup("127.0.0.1:8125/my.prefix", -1)
// or
client := statsd.NewWithPacketSize("127.0.0.1:8125", "my.prefix.", -1)
```

Benchmarks
==========

November 14, 2013:

```
% go test -bench . --benchmem
BenchmarkGaugeNoPrefix             1113 ns/op   80 B/op   4 allocs/op
BenchmarkGaugeWithPrefix           1326 ns/op   80 B/op   4 allocs/op
BenchmarkGaugeNoBuffer             6984 ns/op   80 B/op   4 allocs/op
BenchmarkGaugeWithPrefixNoBuffer   6910 ns/op   80 B/op   4 allocs/op
```

