package sun

import (
	"context"
	"satweave/messenger"
	"satweave/messenger/common"
	task_manager "satweave/shared/task-manager"
	"satweave/utils/logger"
)

// SpecificJobInfo 管理某个具体的 Job 的信息
type SpecificJobInfo struct {
	sourceOps                  map[uint64][]*common.ExecuteTask
	registeredTaskManagerTable *RegisteredTaskManagerTable
	ackMap                     *AcknowledgeTable
}

// NewSpecificJobInfo TODO(qiu): 传的是 task 不是execute task？
func NewSpecificJobInfo(executeTaskMap map[uint64][]*common.ExecuteTask) *SpecificJobInfo {
	return &SpecificJobInfo{
		sourceOps: findSourceOps(executeTaskMap),
		ackMap:    NewAcknowledgeTable(executeTaskMap),
	}
}

func findSourceOps(executeTaskMap map[uint64][]*common.ExecuteTask) map[uint64][]*common.ExecuteTask {
	sourceOps := make(map[uint64][]*common.ExecuteTask)
	for taskManagerId, tasks := range executeTaskMap {
		for _, task := range tasks {
			if len(task.InputEndpoints) == 0 {
				if _, exists := sourceOps[taskManagerId]; !exists {
					sourceOps[taskManagerId] = make([]*common.ExecuteTask, 0)
				}
				sourceOps[taskManagerId] = append(sourceOps[taskManagerId], task)
			}
		}
	}
	return sourceOps
}

func (s *SpecificJobInfo) AcknowledgeCheckpoint(request *AcknowledgeCheckpointRequest) (bool, error) {
	return s.ackMap.acknowledgeCheckpoint(request)
}

func (s *SpecificJobInfo) triggerCheckpoint(checkpointId int64, registeredTaskManagerTable *RegisteredTaskManagerTable,
	cancelJob bool) error {
	// 传入 registered task manager table 是为了找到对应  task manager 的endpoint
	if s.ackMap.hasCheckpoint(checkpointId) {
		logger.Errorf("Checkpoint %v already exists", checkpointId)
		return nil
	}
	for taskManagerId, tasks := range s.sourceOps {
		taskManagerHost := registeredTaskManagerTable.getHost(taskManagerId)
		taskManagerPort := s.registeredTaskManagerTable.getPort(taskManagerId)
		for _, task := range tasks {
			err := s.innerTriggerCheckpoint(taskManagerId, taskManagerHost, taskManagerPort, task.WorkerId,
				task.SubtaskName, checkpointId, cancelJob)
			if err != nil {
				return err
			}
		}
	}
	s.ackMap.registerPendingCheckpoint(checkpointId)
	return nil
}

func (s *SpecificJobInfo) innerTriggerCheckpoint(taskManagerId uint64, subtaskHost string, subtaskPort uint64,
	workerId uint64, subtaskName string, checkpointId int64, cancelJob bool) error {
	conn, err := messenger.GetRpcConn(subtaskHost, subtaskPort)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := task_manager.NewTaskManagerServiceClient(conn)
	logger.Infof("Try to trigger checkpoint for subtask %v(task manager id: %v)", subtaskName, taskManagerId)
	_, err = client.TriggerCheckpoint(context.Background(), &task_manager.TriggerCheckpointRequest{
		WorkerId: workerId,
		Checkpoint: &common.Record_Checkpoint{
			Id:        checkpointId,
			CancelJob: cancelJob,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
