package worker

import (
	"context"
	"satweave/messenger/common"
	"satweave/utils/logger"
	"sync"
	"time"
)

type InputReceiver struct {
	/*
	   PartitionDispenser   \
	   PartitionDispenser   -    subtask   ->   InputReceiver   ->   channel
	   PartitionDispenser   /            (遇到 event 阻塞，类似 Gate)
	*/
	channel    chan *common.Record
	partitions []*InputPartitionReceiver
	// TODO(qiu): barrier 需要 new 并且初始化，大小 = partition 数量
	barrier  *sync.WaitGroup
	allowOne chan struct{}

	ctx context.Context
}

func (i *InputReceiver) RecvData(partitionIdx int64, record *common.Record) {
	i.partitions[partitionIdx].RecvData(record)
}

func (i *InputReceiver) RunAllPartitionReceiver() {
	if i.partitions == nil {
		return
	}
	for _, partition := range i.partitions {
		go partition.innerPraseDataAndCarryToChannel(partition.queue, partition.channel, partition.barrier, partition.allowOne)
	}
}

type InputPartitionReceiver struct {
	ctx      context.Context
	queue    chan *common.Record
	channel  chan *common.Record
	barrier  *sync.WaitGroup
	allowOne chan struct{}
}

func (ipr *InputPartitionReceiver) RecvData(record *common.Record) {
	logger.Infof("push record %v to queue", record)
	ipr.queue <- record
	logger.Infof("push record %v to queue finished", record)
}

func (ipr *InputPartitionReceiver) praseDataAndCarryToChannel(inputQueue chan *common.Record, outputChannel chan *common.Record,
	eventBarrier *sync.WaitGroup, allowOne chan struct{}) error {
	err := ipr.innerPraseDataAndCarryToChannel(inputQueue, outputChannel, eventBarrier, allowOne)
	if err != nil {
		logger.Errorf("InputPartitionReceiver.innerPraseDataAndCarryToChannel() failed: %v", err)
		return err
	}
	return nil
}

func (ipr *InputPartitionReceiver) innerPraseDataAndCarryToChannel(inputQueue chan *common.Record, outputChannel chan *common.Record, barrier *sync.WaitGroup, allowOne chan struct{}) error {
	logger.Infof("start innerPraseDataAndCarryToChannel....")
	needBarrierDataType := []common.DataType{common.DataType_CHECKPOINT, common.DataType_FINISH}
	for {
		select {
		case <-ipr.ctx.Done():
			return nil
		default:
			logger.Infof("block here")
			data := <-inputQueue
			logger.Infof("do not reach here")
			if ContainType(needBarrierDataType, data.DataType) {
				// TODO(qiu): 需要在某个地方初始化 wg
				barrier.Done()
				logger.Infof("Receive Barrier wg --, begin to wait ... ")
				barrier.Wait()
				logger.Infof("Block finished, begin to make checkpoint")
				// 重置 wg，每个协程都只执行一次
				barrier.Add(1)
				// allowOne 用于只让一个协程往 output 发送数据
				select {
				case <-allowOne:
					logger.Infof("Be the allowOne, transfer checkpoint signal")
					outputChannel <- data
					// 休眠 1s
					time.Sleep(1 * time.Second)
					allowOne <- struct{}{}
				default:
					logger.Infof("Do not be the allowOne, do not transfer checkpoint signal")
				}
			} else {
				outputChannel <- data
			}
		}
	}
}

func ContainType(list []common.DataType, element common.DataType) bool {
	for _, v := range list {
		if v == element {
			return true
		}
	}
	return false
}

func (ipr *InputPartitionReceiver) Run() {
	logger.Infof("partition run ...")

	ipr.innerPraseDataAndCarryToChannel(ipr.queue, ipr.channel, ipr.barrier, make(chan struct{}, 1))
}

func NewInputReceiver(ctx context.Context, inputChannel chan *common.Record, inputEndpoints []*common.InputEndpoints) *InputReceiver {
	inputReceiver := &InputReceiver{
		ctx:        ctx,
		channel:    inputChannel,
		barrier:    &sync.WaitGroup{},
		partitions: make([]*InputPartitionReceiver, 0),
		allowOne:   make(chan struct{}, 1),
	}
	for i := 0; i < len(inputEndpoints); i++ {
		inputPartitionReceiver := inputReceiver.NewInputPartitionReceiver(inputReceiver.channel, inputReceiver.barrier, inputReceiver.allowOne)
		inputReceiver.partitions = append(inputReceiver.partitions, inputPartitionReceiver)
	}
	return inputReceiver
}

func (i *InputReceiver) NewInputPartitionReceiver(channel chan *common.Record, eventBarrier *sync.WaitGroup, allowOne chan struct{}) *InputPartitionReceiver {
	return &InputPartitionReceiver{
		ctx:      i.ctx,
		channel:  channel,
		queue:    make(chan *common.Record, 1000),
		barrier:  eventBarrier,
		allowOne: allowOne,
	}
}