package worker

import "satweave/messenger/common"

type InputReceiver struct {
	/*
	   PartitionDispenser   \
	   PartitionDispenser   -    subtask   ->   InputReceiver   ->   channel
	   PartitionDispenser   /            (遇到 event 阻塞，类似 Gate)
	*/
	inputChannel chan interface{}
	partitions   []*InputPartitionReceiver
}

func (i *InputReceiver) RecvData(partitionIdx uint64, record *common.Record) {
	i.partitions[partitionIdx].RecvData(record)
}

type InputPartitionReceiver struct {
	queue chan interface{}
}

func (ipr *InputPartitionReceiver) RecvData(record *common.Record) {
	ipr.queue <- record
}
