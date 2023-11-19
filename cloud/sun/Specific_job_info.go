package sun

import "satweave/messenger/common"

// 管理某个具体的 Job 的信息
type SpecificJobInfo struct {
	sourceOps map[uint64][]*common.ExecuteTask
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

func triggerCheckpoint(registeredTaskManagerTable RegisteredTaskManagerTable) {
	// 传入 registered task manager table 是为了找到对应  task manager 的endpoint

}
