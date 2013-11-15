package statsd

import (
	"github.com/stvp/go-udp-testing"
	"reflect"
	"testing"
	"time"
)

// Return a valid client that sends messages to our go-udp-testing server.
func goodClient(prefix string, packetSize int) Client {
	client, _ := NewWithPacketSize("statsd://localhost:8125/"+prefix, packetSize)
	return client
}

// -- Tests

type parseUrlTestcase struct {
	url    string
	host   string
	prefix string
	good   bool
}

func TestNew(t *testing.T) {
	client, err := New("statsd://broken:9999")
	if err == nil {
		t.Error(err)
	}
	if reflect.TypeOf(client).String() != "*statsd.emptyClient" {
		t.Fatal("A bad connection should return an emptyClient.")
	}

	client, err = New("statsd://localhost:8125")
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(client).String() != "*statsd.statsdClient" {
		t.Fatal("A good connection should return a statsdClient.")
	}
}

func TestGauge(t *testing.T) {
	udp.SetAddr(":8125")
	client := goodClient("", 512)

	// Positive numbers
	udp.ShouldReceiveOnly(t, "bukkit:2|g", func() {
		client.Gauge("bukkit", 2)
		client.Flush()
	})
	// Negative numbers
	udp.ShouldReceiveOnly(t, "bukkit:-12|g", func() {
		client.Gauge("bukkit", -12)
		client.Flush()
	})
	// Large floats
	udp.ShouldReceiveOnly(t, "bukkit:1.234567890123457|g", func() {
		client.Gauge("bukkit", 1.2345678901234568901234)
		client.Flush()
	})
	udp.ShouldReceiveOnly(t, "bukkit:0.000000000000001|g", func() {
		client.Gauge("bukkit", 0.000000000000001)
		client.Flush()
	})
}

func BenchmarkGaugeNoPrefix(b *testing.B) {
	udp.SetAddr(":8125")
	client := goodClient("", 512)

	for i := 0; i < b.N; i++ {
		client.Gauge("metrics.are.cool", 98765.4321)
	}
}

func BenchmarkGaugeWithPrefix(b *testing.B) {
	udp.SetAddr(":8125")
	client := goodClient("some.prefix.here", 512)

	for i := 0; i < b.N; i++ {
		client.Gauge("metrics.are.cool", 98765.4321)
	}
}

func TestCount(t *testing.T) {
	udp.SetAddr(":8125")
	client := goodClient("", 512)

	// Positive numbers
	udp.ShouldReceiveOnly(t, "bukkit:2|c", func() {
		client.Count("bukkit", 2, 1)
		client.Flush()
	})
	// Negative numbers
	udp.ShouldReceiveOnly(t, "bukkit:-10|c", func() {
		client.Count("bukkit", -10, 1)
		client.Flush()
	})
	// Large floats
	udp.ShouldReceiveOnly(t, "bukkit:1.234567890123457|c", func() {
		client.Count("bukkit", 1.2345678901234568901234, 1)
		client.Flush()
	})
	// Sample rates
	udp.ShouldReceiveOnly(t, "bukkit:1|c|@0.999999", func() {
		client.Count("bukkit", 1, 0.999999)
		client.Flush()
	})
}

func TestPrefix(t *testing.T) {
	udp.SetAddr(":8125")
	client := goodClient("dude", 512)

	udp.ShouldReceiveOnly(t, "dude.cool.bukkit:1|c", func() {
		client.Count("cool.bukkit", 1, 1)
		client.Flush()
	})
}

func TestTiming(t *testing.T) {
	udp.SetAddr(":8125")
	client := goodClient("", 512)

	udp.ShouldReceiveOnly(t, "bukkit:250|ms", func() {
		client.Timing("bukkit", 250*time.Millisecond)
		client.Flush()
	})

	udp.ShouldReceiveOnly(t, "bukkit:250000|ms", func() {
		client.Timing("bukkit", 250*time.Second)
		client.Flush()
	})
}

func TestCountUnique(t *testing.T) {
	udp.SetAddr(":8125")
	client := goodClient("", 512)

	udp.ShouldReceiveOnly(t, "bukkit:foo|s", func() {
		client.CountUnique("bukkit", "foo")
		client.Flush()
	})

	udp.ShouldReceiveOnly(t, "bukkit:foo_bar_1_baz_biz|s", func() {
		client.CountUnique("bukkit", "foo:bar -1- baz|biz")
		client.Flush()
	})
}

func TestFloatFormatting(t *testing.T) {
	udp.SetAddr(":8125")
	client := goodClient("", 512)

	udp.ShouldReceiveOnly(t, "foo:0.0000000000667428|g", func() {
		client.Gauge("foo", 6.67428e-11)
		client.Flush()
	})

	udp.ShouldReceiveOnly(t, "foo:1.2345678901234567|g", func() {
		client.Gauge("foo", 1.234567890123456789)
		client.Flush()
	})

	udp.ShouldReceiveOnly(t, "foo:1234567000000000000|g", func() {
		client.Gauge("foo", 1234567000000000000)
		client.Flush()
	})
}

func TestBuffer(t *testing.T) {
	udp.SetAddr(":8125")

	udp.ShouldNotReceive(t, "b:2|c", func() {
		client := goodClient("", 512)
		client.Count("a", 1, 1)
		client.Flush()
		client.Count("b", 2, 1)
		client.Count("c", 3, 1)
	})

	udp.ShouldReceiveOnly(t, "a:1|c\nb:2|c\nc:3|c", func() {
		client := goodClient("", 512)
		client.Count("a", 1, 1)
		client.Count("b", 2, 1)
		client.Count("c", 3, 1)
		client.Flush()
	})

	truncatedPacket := `four.score.and.seven.years.ago:0|c
four.score.and.seven.years.ago:1|c
four.score.and.seven.years.ago:2|c
four.score.and.seven.years.ago:3|c
four.score.and.seven.years.ago:4|c
four.score.and.seven.years.ago:5|c
four.score.and.seven.years.ago:6|c
four.score.and.seven.years.ago:7|c
four.score.and.seven.years.ago:8|c
four.score.and.seven.years.ago:9|c
four.score.and.seven.years.ago:10|c
four.score.and.seven.years.ago:11|c
four.score.and.seven.years.ago:12|c
four.score.and.seven.years.ago:13|c`
	udp.ShouldReceiveOnly(t, truncatedPacket, func() {
		client := goodClient("", 512)
		for i := 0; i < 16; i++ {
			client.Count("four.score.and.seven.years.ago", float64(i), 1)
		}
		client.Flush()
	})

	fullPacket := `four.score.and.seven.years.ago:0|c
four.score.and.seven.years.ago:1|c
four.score.and.seven.years.ago:2|c
four.score.and.seven.years.ago:3|c
four.score.and.seven.years.ago:4|c
four.score.and.seven.years.ago:5|c
four.score.and.seven.years.ago:6|c
four.score.and.seven.years.ago:7|c
four.score.and.seven.years.ago:8|c
four.score.and.seven.years.ago:9|c
four.score.and.seven.years.ago:10|c
four.score.and.seven.years.ago:11|c
four.score.and.seven.years.ago:12|c
four.score.and.seven.years.ago:13|c
four.score.and.seven.years.ago:14|c
four.score.and.seven.years.ago:15|c`
	udp.ShouldReceiveOnly(t, fullPacket, func() {
		client := goodClient("", 1024)
		for i := 0; i < 16; i++ {
			client.Count("four.score.and.seven.years.ago", float64(i), 1)
		}
		client.Flush()
	})

	udp.ShouldReceiveOnly(t, "a:1|c", func() {
		client := goodClient("", 0)
		client.Count("a", 1, 1)
		// No flush needed
	})
}
