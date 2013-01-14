package statsd

import (
	"net"
	"reflect"
	"testing"
	"time"
)

//
// Helpers
//

var addr, _ = net.ResolveUDPAddr("udp", ":8125")
var listener, _ = net.ListenUDP("udp", addr)

type fn func()

// Expect to receive the given message on UDP port 8125 while running the given
// func.
func expectMessage(t *testing.T, expected string, body fn) {
	result := make(chan string)
	go func() {
		message := make([]byte, 512)
		n, _, err := listener.ReadFrom(message)
		if err != nil {
			t.Fatal(err)
		}
		result <- string(message[0:n])
	}()
	body()
	got := <-result
	if got != expected {
		t.Errorf("Expected %#v but got %#v instead.", expected, got)
	}
}

// Return a valid client that sends messages to the server set up above.
func goodClient(namespace string) StatsReporter {
	client, _ := New("localhost:8125", namespace)
	return client
}

//
// The tests
//

func TestBadConnection(t *testing.T) {
	client, err := New("broken:9999", "")
	if err == nil {
		t.Error(err)
	}
	if reflect.TypeOf(client).String() != "*statsd.emptyClient" {
		t.Fatal("A bad connection should return an emptyClient.")
	}
}

func TestGoodConnection(t *testing.T) {
	client, err := New("localhost:8125", "")
	if err != nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(client).String() != "*statsd.statsdClient" {
		t.Fatal("A good connection should return a statsdClient.")
	}
}

func TestIncrement(t *testing.T) {
	expectMessage(t, "bukkit:1|c", func() {
		goodClient("").Increment("bukkit", 1)
	})
}

func TestIncrementNamespace(t *testing.T) {
	expectMessage(t, "dude.cool.bukkit:1|c", func() {
		goodClient("dude").Increment("cool.bukkit", 1)
	})
}

// TODO: How can we stub rand.Float32()
func TestIncrementSampleRate(t *testing.T) {
	expectMessage(t, "bukkit:1|c|@0.999999", func() {
		goodClient("").Increment("bukkit", 0.999999)
	})
}

func TestDecrement(t *testing.T) {
	expectMessage(t, "bukkit:-1|c", func() {
		goodClient("").Decrement("bukkit", 1.0)
	})
}

func TestTiming(t *testing.T) {
	expectMessage(t, "bukkit:250|ms", func() {
		goodClient("").Timing("bukkit", 250 * time.Millisecond)
	})
	expectMessage(t, "bukkit:250000|ms", func() {
		goodClient("").Timing("bukkit", 250 * time.Second)
	})
}

