package worker

import (
	"context"
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

	// channel = inputChannel 与 worker 共用
	InputChannel       chan *common.Record
	PartitionReceivers []*InputPartitionReceiver
	// TODO(qiu): barrier 需要 new 并且初始化，大小 = partition 数量
	EventBarrier *sync.WaitGroup
	AllowOne     chan struct{}

	ctx context.Context
}

func (i *InputReceiver) RecvData(partitionIdx int64, record *common.Record) {
	i.PartitionReceivers[partitionIdx].RecvData(record)
}

func (i *InputReceiver) RunAllPartitionReceiver() {
	if i.PartitionReceivers == nil {
		logger.Errorf("partition receivers is nil")
		return
	}
	for _, partitionReceiver := range i.PartitionReceivers {
		go partitionReceiver.Run()
	}
}

type InputPartitionReceiver struct {
	ctx          context.Context
	InnerQueue   chan *common.Record
	InputChannel chan *common.Record
	EventBarrier *sync.WaitGroup
	AllowOne     chan struct{}
}

// TODO 1.30 看一下这个函数的调用关系
func (ipr *InputPartitionReceiver) RecvData(record *common.Record) {
	logger.Infof("push record %v to InputPartitionReceiver Inner Queue", record)
	ipr.InnerQueue <- record
	logger.Infof("push record %v to InputPartitionReceiver Inner Queue", record)
}

// PhraseDataAndCarryToChannel 用于解析 InnerQueue 中接收到的数据，并且写入 InputChannel
func (ipr *InputPartitionReceiver) PhraseDataAndCarryToChannel() {
	logger.Infof("start PhraseDataAndCarryToChannel....")
	needBarrierDataType := []common.DataType{common.DataType_CHECKPOINT, common.DataType_FINISH}
	for {
		select {
		case <-ipr.ctx.Done():
			logger.Infof("context done, close input partition receiver")
			return
		default:
			// 从 inputQueue 中读取数据
			record := <-ipr.InnerQueue
			// 如果是Barrier的数据类型，需要阻塞
			if ContainType(needBarrierDataType, record.DataType) {
				ipr.EventBarrier.Done()
				logger.Infof("Receive Barrier wg --, begin to wait ... ")
				ipr.EventBarrier.Wait()
				logger.Infof("Block finished, begin to make checkpoint")
				// 重置 wg，每个协程都只执行一次
				ipr.EventBarrier.Add(1)
				// allowOne 用于只让一个协程往 output 发送数据
				select {
				case <-ipr.AllowOne:
					logger.Infof("Be the allowOne, transfer checkpoint or finish signal")
					ipr.InputChannel <- record
					//// TODO 一秒会不会有点久
					//time.Sleep(100 * time.Second)
					ipr.AllowOne <- struct{}{}
				default:
				}
			} else {
				ipr.InputChannel <- record
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
	ipr.PhraseDataAndCarryToChannel()
}

func NewInputReceiver(ctx context.Context, inputChannel chan *common.Record, inputEndpoints []*common.InputEndpoints) *InputReceiver {
	inputReceiver := &InputReceiver{
		ctx:                ctx,
		InputChannel:       inputChannel,
		EventBarrier:       &sync.WaitGroup{},
		PartitionReceivers: make([]*InputPartitionReceiver, 0),
		AllowOne:           make(chan struct{}, 1),
	}
	// 在这里初始化 barrier
	for i := 0; i < len(inputEndpoints); i++ {
		inputReceiver.EventBarrier.Add(1)
	}

	// 初始化 PartitionReceivers
	for i := 0; i < len(inputEndpoints); i++ {
		partitionReceiver := inputReceiver.NewInputPartitionReceiver(inputReceiver.InputChannel, inputReceiver.EventBarrier, inputReceiver.AllowOne)
		inputReceiver.PartitionReceivers = append(inputReceiver.PartitionReceivers, partitionReceiver)
	}
	return inputReceiver
}

func (i *InputReceiver) NewInputPartitionReceiver(inputChannel chan *common.Record, eventBarrier *sync.WaitGroup,
	allowOne chan struct{}) *InputPartitionReceiver {
	return &InputPartitionReceiver{
		ctx:          i.ctx,
		InputChannel: inputChannel,
		// TODO(qiu): 这里可以调整 InnerQueue 的大小
		InnerQueue:   make(chan *common.Record, 1000),
		EventBarrier: eventBarrier,
		AllowOne:     allowOne,
	}
}
