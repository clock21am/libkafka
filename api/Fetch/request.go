package Fetch

import (
	"github.com/mkocikowski/libkafka/api"
)

func NewRequest(topic string, partition int32, offset int64) *api.Request {
	p := Partition{
		Partition:         partition,
		FetchOffset:       offset,
		PartitionMaxBytes: 100 << 20,
	}
	t := Topic{
		Topic:      topic,
		Partitions: []Partition{p, p},
	}
	return &api.Request{
		ApiKey:        api.Fetch,
		ApiVersion:    6,
		CorrelationId: 0,
		ClientId:      "",
		Body: Request{
			ReplicaId:     -1,
			MaxWaitTimeMs: 1000,
			MinBytes:      10 << 20,
			MaxBytes:      100 << 20,
			Topics:        []Topic{t},
		},
	}
}

type Request struct {
	ReplicaId      int32
	MaxWaitTimeMs  int32
	MinBytes       int32
	MaxBytes       int32
	IsolationLevel int8 // not used
	Topics         []Topic
}

type Topic struct {
	Topic      string
	Partitions []Partition
}

type Partition struct {
	Partition         int32
	FetchOffset       int64
	LogStartOffset    int64 // not used
	PartitionMaxBytes int32
}
