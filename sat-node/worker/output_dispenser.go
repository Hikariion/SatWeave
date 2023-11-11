package worker

import (
	"context"
	"satweave/messenger"
	"satweave/messenger/common"
	task_manager "satweave/shared/task-manager"
	"satweave/utils/logger"
)

/*
OutputDispenser

	                                           / PartitionDispenser
	SubTask   ->   channel   ->   Dispenser    - PartitionDispenser   ->  SubTask
	                           (SubTaskClient) \ PartitionDispenser
*/
type OutputDispenser struct {
	channel                  chan *common.Record
	outPutEndPoints          []*common.OutputEndpoints
	subtaskName              string
	partitionIdx             uint64
	outputPartitionDispenser []*OutputPartitionDispenser
}

func (o *OutputDispenser) pushData(record *common.Record) {
	o.channel <- record
}

func (o *OutputDispenser) partitioningDataAndCarryToNextSubtask(inputChannel chan *common.Record, outputEndpoints []*common.OutputEndpoints,
	subtaskName string, partitionIdx uint64) {
	partitions := make([]*OutputPartitionDispenser, 0)
	for _, endPoint := range outputEndpoints {
		partitions = append(partitions, NewOutputPartitionDispenser(subtaskName, partitionIdx, endPoint))
	}

	partitionNum := len(partitions)

	needBroadCastDataType := common.DataType_CHECKPOINT

	for {
		// TODO(qiu): 这个inputchannel是不是应该从结构体里读？
		record := <-inputChannel
		if record.DataType == needBroadCastDataType {
			// BroadCast
			for _, outputPartitionDispenser := range partitions {
				outputPartitionDispenser.pushData(record)
			}
		} else {
			// Partitioning
			partitionIdx := -1
			if record.PartitionKey != nil {

			}
		}
	}
}

func NewOutputDispenser(outputChannel chan *common.Record, outputEndpoints []*common.OutputEndpoints, subtaskName string, partitionIdx uint64) *OutputDispenser {
	// 这里 output_endpoints 是按 partition_idx 顺序排列的
	outputDispenser := &OutputDispenser{
		channel:                  outputChannel,
		outPutEndPoints:          outputEndpoints,
		subtaskName:              subtaskName,
		partitionIdx:             partitionIdx,
		outputPartitionDispenser: nil,
	}
	return outputDispenser
}

type OutputPartitionDispenser struct {
	subTaskName  string
	partitionIdx uint64
	endPoint     *common.OutputEndpoints
}

func (opd *OutputPartitionDispenser) pushData(record *common.Record) error {
	// TODO： use rpc to push data need get
	// 获得 endpoint 对应的 taskManager 的 rpc client
	taskManagetHost := opd.endPoint.Host
	taskManagerPort := opd.endPoint.Port
	conn, err := messenger.GetRpcConn(taskManagetHost, taskManagerPort)
	if err != nil {
		logger.Errorf("get rpc conn failed: %v", err)
		return err
	}
	client := task_manager.NewTaskManagerServiceClient(conn)
	// record 里需要标识 workerId
	workerId := opd.endPoint.WorkerId
	request := &task_manager.PushRecordRequest{
		WorkerId:     workerId,
		Record:       record,
		PartitionIdx: opd.partitionIdx,
		FromSubtask:  opd.subTaskName,
	}
	// 调用 pushRecord
	_, err = client.PushRecord(context.Background(), request)
	if err != nil {
		logger.Errorf("push record to task manager failed: %v", err)
		return err
	}
	return nil
}

func NewOutputPartitionDispenser(subTaskName string, partitionIdx uint64, endPoint *common.OutputEndpoints) *OutputPartitionDispenser {
	return &OutputPartitionDispenser{
		subTaskName:  subTaskName,
		partitionIdx: partitionIdx,
		endPoint:     endPoint,
	}
}
