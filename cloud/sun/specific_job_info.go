package sun

import (
	"context"
	"satweave/messenger"
	"satweave/messenger/common"
	task_manager "satweave/shared/task-manager"
)

// 管理某个具体的 Job 的信息
type SpecificJobInfo struct {
	sourceOps                  map[uint64][]*common.ExecuteTask
	registeredTaskManagerTable *RegisteredTaskManagerTable
}

// NewSpecificJobInfo TODO(qiu): 传的是 task 不是execute task？
func NewSpecificJobInfo(executeTaskMap map[uint64][]*common.ExecuteTask) *SpecificJobInfo {
	return &SpecificJobInfo{
		sourceOps: findSourceOps(executeTaskMap),
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

func (s *SpecificJobInfo) triggerCheckpoint(registeredTaskManagerTable *RegisteredTaskManagerTable) error {
	// 传入 registered task manager table 是为了找到对应  task manager 的endpoint
	for taskManagerId, tasks := range s.sourceOps {
		taskManagerHost := s.registeredTaskManagerTable.getHost(taskManagerId)
		port := s.registeredTaskManagerTable.getPort(taskManagerId)
		for _, task := range tasks {
			subtaskName := task.SubtaskName
			// workerId := task.WorkerId

			// 获取 rpc client
			conn, err := messenger.GetRpcConn(taskManagerHost, port)
			if err != nil {
				return err
			}
			client := task_manager.NewTaskManagerServiceClient(conn)
			// 触发 checkpoint
			//TODO: 这里还没写完
			_, err = client.TriggerCheckpoint(context.Background(), &task_manager.TriggerCheckpointRequest{})

			conn.Close()
		}
	}
	return nil
}
