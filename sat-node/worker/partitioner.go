package worker

import (
	"math/rand"
	"satweave/messenger/common"
	"satweave/utils/logger"
	"strconv"
	"time"
)

type PartitionerBase interface {
	Partitioning(record *common.Record, partitionNum uint64) uint64
}

type KeyPartitioner struct {
}

func (k *KeyPartitioner) Partitioning(record *common.Record, partitionNum uint64) (uint64, error) {
	partitionKey := record.PartitionKey
	// 把 partitionKey 转化为 uint64
	p, err := strconv.ParseInt(partitionKey, 10, 64)
	if err != nil {
		logger.Errorf("KeyPartitioner.Partitioning() failed: %v", err)
		return 0, err
	}
	return uint64(p) % partitionNum, nil
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
