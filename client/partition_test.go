package client

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/mkocikowski/libkafka/api/Metadata"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestIntergationPartitionClientSuccess(t *testing.T) {
	bootstrap := "localhost:9092"
	topic := fmt.Sprintf("test-%x", rand.Uint32())
	if _, err := CallCreateTopic(bootstrap, topic, 1, 1); err != nil {
		t.Fatal(err)
	}
	c := &PartitionClient{
		Bootstrap: bootstrap,
		Topic:     topic,
		Partition: 0,
	}
	r, err := c.ListOffsets(0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", r)
}

func TestIntergationPartitionClientBadBootstrap(t *testing.T) {
	bootstrap := "foo"
	topic := fmt.Sprintf("test-%x", rand.Uint32()) // do not create
	c := &PartitionClient{
		Bootstrap: bootstrap,
		Topic:     topic,
		Partition: 0,
	}
	_, err := c.ListOffsets(0)
	if err == nil {
		t.Fatal("expected 'dial tcp' error")
	}
	t.Log(err)
}

func TestIntergationPartitionClientTopicDoesNotExist(t *testing.T) {
	bootstrap := "localhost:9092"
	topic := fmt.Sprintf("test-%x", rand.Uint32()) // do not create
	c := &PartitionClient{
		Bootstrap: bootstrap,
		Topic:     topic,
		Partition: 0,
	}
	_, err := c.ListOffsets(0)
	if !errors.Is(err, ErrPartitionDoesNotExist) {
		t.Fatal(err)
	}
	t.Log(err)
}

// the purpose of this test is to test that when ConnMaxIdle is set connection
// is automatically closed and reopened when the idle time is exceeded
func TestIntergationPartitionClientConnectionIdleTimeout(t *testing.T) {
	bootstrap := "localhost:9092"
	topic := fmt.Sprintf("test-%x", rand.Uint32())
	if _, err := CallCreateTopic(bootstrap, topic, 1, 1); err != nil {
		t.Fatal(err)
	}
	timeout := 50 * time.Millisecond
	c := &PartitionClient{
		Bootstrap:   bootstrap,
		Topic:       topic,
		Partition:   0,
		ConnMaxIdle: timeout,
	}
	// make first call to open connection
	if _, err := c.ListOffsets(0); err != nil {
		t.Fatal(err)
	}
	// record the connection
	conn := c.Conn()
	// make second call
	if _, err := c.ListOffsets(0); err != nil {
		t.Fatal(err)
	}
	// ensure the connection is the same connection
	if c.Conn() != conn {
		t.Fatal("different connection")
	}
	// now exceed the timeout
	time.Sleep(timeout)
	// third call
	if _, err := c.ListOffsets(0); err != nil {
		t.Fatal(err)
	}
	// now there should be different connection
	if c.Conn() == conn {
		t.Fatal("same connection")
	}
}

func TestUnitLeaderString(t *testing.T) {
	b := &Metadata.Broker{Rack: "foo", NodeId: 1, Host: "bar", Port: 9092}
	s := fmt.Sprintf("%v", b)
	if s != "foo:1:bar:9092" {
		t.Fatal(s)
	}
}
