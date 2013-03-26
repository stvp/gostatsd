package statsd

import (
	"net"
	"reflect"
	"testing"
	"time"
)

// -- Helpers

var addr, _ = net.ResolveUDPAddr("udp", ":8125")
var listener, _ = net.ListenUDP("udp", addr)

type fn func()

// Expect to receive the given message on UDP port 8125 while running the given
// func.
func expectMessage(t *testing.T, expected string, body fn) {
	got := getMessage(body)

	// Check the actual received bytes
	if got != expected {
		t.Errorf("Expected %#v but got %#v instead.", expected, got)
	}
}

// Get the UDP data received on port 8125 while running the given function.
func getMessage(body fn) string {
	result := make(chan string)
	go func() {
		message := make([]byte, 1024)
		n, _, _ := listener.ReadFrom(message)
		result <- string(message[0:n])
	}()

	body()

	return <-result
}

// Return a valid client that sends messages to the server set up above.
func goodClient(prefix string) StatsReporter {
	client, _ := New("localhost:8125", prefix)
	return client
}

// -- Tests

func TestNew(t *testing.T) {
	client, err := New("broken:9999", "")
	if err == nil {
		t.Error(err)
	}
	if reflect.TypeOf(client).String() != "*statsd.emptyClient" {
		t.Fatal("A bad connection should return an emptyClient.")
	}

	client, err = New("localhost:8125", "")
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(client).String() != "*statsd.statsdClient" {
		t.Fatal("A good connection should return a statsdClient.")
	}
}

func TestGauge(t *testing.T) {
	// Positive numbers
	expectMessage(t, "bukkit:2|g", func() {
		client := goodClient("")
		client.Gauge("bukkit", 2)
		client.Flush()
	})
	// Negative numbers
	expectMessage(t, "bukkit:-12|g", func() {
		client := goodClient("")
		client.Gauge("bukkit", -12)
		client.Flush()
	})
	// Large floats
	expectMessage(t, "bukkit:1.234567890123457|g", func() {
		client := goodClient("")
		client.Gauge("bukkit", 1.2345678901234568901234)
		client.Flush()
	})
}

func TestCount(t *testing.T) {
	// Positive numbers
	expectMessage(t, "bukkit:2|c", func() {
		client := goodClient("")
		client.Count("bukkit", 2, 1)
		client.Flush()
	})
	// Negative numbers
	expectMessage(t, "bukkit:-10|c", func() {
		client := goodClient("")
		client.Count("bukkit", -10, 1)
		client.Flush()
	})
	// Large floats
	expectMessage(t, "bukkit:1.234567890123457|c", func() {
		client := goodClient("")
		client.Count("bukkit", 1.2345678901234568901234, 1)
		client.Flush()
	})
	// Sample rates
	expectMessage(t, "bukkit:1|c|@0.999999", func() {
		client := goodClient("")
		client.Count("bukkit", 1, 0.999999)
		client.Flush()
	})
}

func TestPrefix(t *testing.T) {
	expectMessage(t, "dude.cool.bukkit:1|c", func() {
		client := goodClient("dude.")
		client.Count("cool.bukkit", 1, 1)
		client.Flush()
	})
}

func TestTiming(t *testing.T) {
	expectMessage(t, "bukkit:250|ms", func() {
		client := goodClient("")
		client.Timing("bukkit", 250*time.Millisecond)
		client.Flush()
	})
	expectMessage(t, "bukkit:250000|ms", func() {
		client := goodClient("")
		client.Timing("bukkit", 250*time.Second)
		client.Flush()
	})
}

func TestCountUnique(t *testing.T) {
	expectMessage(t, "bukkit:foo|s", func() {
		client := goodClient("")
		client.CountUnique("bukkit", "foo")
		client.Flush()
	})
	expectMessage(t, "bukkit:foo_bar_1_baz_biz|s", func() {
		client := goodClient("")
		client.CountUnique("bukkit", "foo:bar -1- baz|biz")
		client.Flush()
	})
}

func TestBuffer(t *testing.T) {
	expectMessage(t, "a:1|c\nb:2|c\nc:3|c", func() {
		client := goodClient("")
		client.Count("a", 1, 1)
		client.Count("b", 2, 1)
		client.Count("c", 3, 1)
		client.Flush()
	})

	largePacket := `four.score.and.seven.years.ago:0|c
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
	message := getMessage(func() {
		client := goodClient("")
		for i := 0; i < 16; i++ {
			client.Count("four.score.and.seven.years.ago", float64(i), 1)
		}
		client.Flush()
	})
	if message != largePacket {
		t.Errorf("Expected %d bytes, but got %d", len(largePacket), len(message))
		t.Error(message)
	}
}
