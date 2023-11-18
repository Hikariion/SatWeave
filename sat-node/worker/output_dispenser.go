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
	ctx                      context.Context
	channel                  chan *common.Record
	outPutEndPoints          []*common.OutputEndpoints
	subtaskName              string
	partitionIdx             int64
	outputPartitionDispenser []*OutputPartitionDispenser
}

func (o *OutputDispenser) pushData(record *common.Record) {
	o.channel <- record
}

func (o *OutputDispenser) partitioningDataAndCarryToNextSubtask(inputChannel chan *common.Record, outputEndPoints []*common.OutputEndpoints,
	subtaskName string, partitionIdx int64) error {

	go o.innerPartitioningDataAndCarryToNextSubtask(inputChannel, outputEndPoints, subtaskName, partitionIdx)

	return nil
}

func (o *OutputDispenser) innerPartitioningDataAndCarryToNextSubtask(inputChannel chan *common.Record, outputEndpoints []*common.OutputEndpoints,
	subtaskName string, partitionIdx int64) error {

	// 初始化 outputPartitionDispenser
	partitions := make([]*OutputPartitionDispenser, 0)
	for _, endPoint := range outputEndpoints {
		partitions = append(partitions, NewOutputPartitionDispenser(subtaskName, partitionIdx, endPoint))
	}

	partitionNum := len(partitions)

	needBroadCastDataType := []common.DataType{common.DataType_CHECKPOINT, common.DataType_FINISH}

	// TODO(qiu): 利用 error break？
	for {
		select {
		case <-o.ctx.Done():
			return nil
		default:
			logger.Infof("output dispenser %s start, block here", o.subtaskName)
			record := <-inputChannel
			logger.Infof("output dispenser %s get record: %v", o.subtaskName, record)
			if ContainType(needBroadCastDataType, record.DataType) {
				// BroadCast
				for _, outputPartitionDispenser := range partitions {
					err := outputPartitionDispenser.pushData(record)
					if err != nil {
						logger.Errorf("push data to next subtask failed: %v", err)
					}
				}
			} else {
				// Partitioning
				var partitionIdx int64
				partitionIdx = -1
				if record.PartitionKey != -1 {
					keyPartitioner := &KeyPartitioner{}
					idx, err := keyPartitioner.Partitioning(record, int64(partitionNum))
					if err != nil {
						logger.Errorf("partitioning record failed: %v", err)
						return err
					}
					partitionIdx = idx
				} else {
					randomPartitioner := &RandomPartitioner{}
					idx, err := randomPartitioner.Partitioning(record, int64(partitionNum))
					if err != nil {
						logger.Errorf("partitioning record failed: %v", err)
						return err
					}
					partitionIdx = idx
				}
				err := partitions[partitionIdx].pushData(record)
				if err != nil {
					logger.Errorf("push data to next subtask failed: %v", err)
					return err
				}
			}
		}
	}
}

func (o *OutputDispenser) Run() {
	err := o.partitioningDataAndCarryToNextSubtask(o.channel, o.outPutEndPoints, o.subtaskName, o.partitionIdx)
	if err != nil {
		logger.Errorf("partitioning data and carry to next subtask failed: %v", err)
		return
	}
}

func NewOutputDispenser(ctx context.Context, outputChannel chan *common.Record, outputEndpoints []*common.OutputEndpoints, subtaskName string, partitionIdx int64) *OutputDispenser {
	// 这里 output_endpoints 是按 partition_idx 顺序排列的
	outputDispenser := &OutputDispenser{
		ctx:                      ctx,
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
	partitionIdx int64
	endPoint     *common.OutputEndpoints
}

// 把 record 发送到下一个节点
func (opd *OutputPartitionDispenser) pushData(record *common.Record) error {
	// TODO： use rpc to push data need get
	// 获得 endpoint 对应的 taskManager 的 rpc client
	taskManagerHost := opd.endPoint.Host
	taskManagerPort := opd.endPoint.Port
	conn, err := messenger.GetRpcConn(taskManagerHost, taskManagerPort)
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

func NewOutputPartitionDispenser(subTaskName string, partitionIdx int64, endPoint *common.OutputEndpoints) *OutputPartitionDispenser {
	return &OutputPartitionDispenser{
		subTaskName:  subTaskName,
		partitionIdx: partitionIdx,
		endPoint:     endPoint,
	}
}
