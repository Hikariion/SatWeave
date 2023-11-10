package worker

import (
	"math/rand"
	"satweave/messenger/common"
	"time"
)

type PartitionerBase interface {
	Partitioning(record *common.Record, partitionNum uint64) uint64
}

type KeyPartitioner struct {
}

func (k *KeyPartitioner) Partitioning(record *common.Record, partitionNum uint64) (uint64, error) {
	partitionKey := record.PartitionKey
	return partitionKey % partitionNum, nil
}

type TimestampPartitioner struct {
}

func (t *TimestampPartitioner) Partitioning(record *common.Record, partitionNum uint64) (uint64, error) {
	timestamp := record.Timestamp
	return timestamp % partitionNum, nil
}

type RandomPartitioner struct {
}

func (r *RandomPartitioner) Partitioning(record *common.Record, partitionNum uint64) (uint64, error) {
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(int(partitionNum))
	return uint64(randomNum) % partitionNum, nil
}
