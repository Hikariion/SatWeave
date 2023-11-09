package worker

import "satweave/messenger/common"

/*
OutputDispenser

	                                           / PartitionDispenser
	SubTask   ->   channel   ->   Dispenser    - PartitionDispenser   ->  SubTask
	                           (SubTaskClient) \ PartitionDispenser
*/
type OutputDispenser struct {
	channel         chan *common.Record
	outPutEndPoints []*common.OutputEndpoints
	subtaskName     string
	partitionIdx    uint64
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
		channel:         outputChannel,
		outPutEndPoints: outputEndpoints,
		subtaskName:     subtaskName,
		partitionIdx:    partitionIdx,
	}
	return outputDispenser
}

type OutputPartitionDispenser struct {
	subTaskName  string
	partitionIdx uint64
	endPoint     *common.OutputEndpoints
}

func (opd *OutputPartitionDispenser) pushData(record *common.Record) {
	// TODO： use rpc to push data need get
}

func NewOutputPartitionDispenser(subTaskName string, partitionIdx uint64, endPoint *common.OutputEndpoints) *OutputPartitionDispenser {
	return &OutputPartitionDispenser{
		subTaskName:  subTaskName,
		partitionIdx: partitionIdx,
		endPoint:     endPoint,
	}
}
