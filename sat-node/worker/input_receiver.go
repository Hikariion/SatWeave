package worker

import (
	"satweave/messenger/common"
	"satweave/utils/logger"
	"sync"
)

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
	queue chan *common.Record
}

func (ipr *InputPartitionReceiver) RecvData(record *common.Record) {
	ipr.queue <- record
}

func praseDataAndCarryToChannel(inputQueue chan *common.Record, outputChannel chan *common.Record, wg *sync.WaitGroup, allowOne chan struct{}) error {
	var needBarrierDataType = common.DataType_CHECKPOINT
	for {
		data := <-inputQueue
		if data.DataType == needBarrierDataType {
			// TODO(qiu): 需要在某个地方初始化 wg
			wg.Done()
			logger.Infof("Receive Barrier wg --, begin to wait ... ")
			wg.Wait()
			logger.Infof("Block finished, begin to make checkpoint")
			// 重置 wg，每个协程都只执行一次
			wg.Add(1)
			// allowOne 用于只让一个协程往 output 发送数据
			select {
			case <-allowOne:
				logger.Infof("Be the allowOne, transfer checkpoint signal")
				outputChannel <- data
				allowOne <- struct{}{}
			default:
				logger.Infof("Do not be the allowOne, do not transfer checkpoint signal")
			}
		} else {
			outputChannel <- data
		}
	}
}
