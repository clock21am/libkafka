package producer

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/mkocikowski/libkafka/api/Produce"
	"github.com/mkocikowski/libkafka/batch"
	"github.com/mkocikowski/libkafka/client"
	"github.com/mkocikowski/libkafka/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestIntergationPartitionProducer(t *testing.T) {
	bootstrap := "localhost:9092"
	topic := fmt.Sprintf("test-%x", rand.Uint32())
	if _, err := client.CallCreateTopic(bootstrap, topic, 1, 1); err != nil {
		t.Fatal(err)
	}
	p := &PartitionProducer{
		PartitionClient: client.PartitionClient{
			Bootstrap: bootstrap,
			Topic:     topic,
			Partition: 0,
		},
		Acks:      1,
		TimeoutMs: 1000,
	}
	if _, err := p.ProduceStrings(time.Now(), "foo", "bar"); err != nil {
		t.Fatal(err)
	}
	resp, err := p.ProduceStrings(time.Now(), "monkey", "banana")
	if err != nil {
		t.Fatal(err)
	}
	if resp.BaseOffset != 2 {
		t.Fatal(resp.BaseOffset)
	}
	if _, err := p.ProduceStrings(time.Now(), []string{}...); err != batch.ErrEmpty {
		t.Fatal(err)
	}
	p.Partition = 1
	if resp, _ := p.ProduceStrings(time.Now(), "hello"); resp.ErrorCode != errors.UNKNOWN_TOPIC_OR_PARTITION {
		t.Fatal(errors.Descriptions[int(resp.ErrorCode)])
	}
}

func TestIntergationPartitionProducerSingleBatch(t *testing.T) {
	bootstrap := "localhost:9092"
	topic := fmt.Sprintf("test-%x", rand.Uint32())
	if _, err := client.CallCreateTopic(bootstrap, topic, 1, 1); err != nil {
		t.Fatal(err)
	}
	p := &PartitionProducer{
		PartitionClient: client.PartitionClient{
			Bootstrap: bootstrap,
			Topic:     topic,
			Partition: 0,
		},
		Acks:      1,
		TimeoutMs: 1000,
	}
	now := time.Unix(1584485804, 0)
	b, _ := batch.NewBuilder(now).AddStrings("foo", "bar").Build(now)
	if b.Crc != 0 {
		t.Fatal(b.Crc)
	}
	resp, err := p.Produce(b)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrorCode != errors.NONE {
		t.Fatal(resp.ErrorCode)
	}
	if b.Crc != 3094838044 {
		t.Fatal(b.Crc)
	}
	t.Logf("%+v", resp)
	//
	p.Acks = 2
	resp, err = p.Produce(b)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrorCode != errors.INVALID_REQUIRED_ACKS {
		t.Fatalf("%+v", resp)
	}
}

func TestIntergationPartitionProducerBadTopic(t *testing.T) {
	p := &PartitionProducer{
		PartitionClient: client.PartitionClient{
			Bootstrap: "localhost:9092",
			Topic:     "no-such-topic",
		},
	}
	resp, err := p.ProduceStrings(time.Now(), "foo", "bar")
	if err == nil {
		t.Fatalf("%+v", resp)
	}
	t.Log(err)
}

func TestIntergationPartitionProducerCorruptBytes(t *testing.T) {
	bootstrap := "localhost:9092"
	topic := fmt.Sprintf("test-%x", rand.Uint32())
	if _, err := client.CallCreateTopic(bootstrap, topic, 1, 1); err != nil {
		t.Fatal(err)
	}
	p := &PartitionProducer{
		PartitionClient: client.PartitionClient{
			Bootstrap: bootstrap,
			Topic:     topic,
			Partition: 0,
		},
	}
	now := time.Unix(1584485804, 0)
	b, _ := batch.NewBuilder(now).AddStrings("foo", "bar").Build(now)
	corrupted := b.Marshal()
	corrupted[len(corrupted)-1] = math.MaxUint8 - corrupted[len(corrupted)-1]
	// calling PartitionClient.Produce and not just Produce so that batch is not re-marshaled
	args := &Produce.Args{
		Topic:     topic,
		Partition: 0,
		Acks:      1,
		TimeoutMs: 1000,
	}
	resp, err := p.PartitionClient.Produce(args, corrupted)
	if err != nil {
		t.Fatal(err)
	}
	if r, _ := parseResponse(resp); r.ErrorCode != errors.CORRUPT_MESSAGE {
		t.Fatalf("%+v", r)
	}
}
