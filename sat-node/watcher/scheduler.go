package watcher

import (
	"satweave/sat-node/task-manager"
	"satweave/utils/logger"
)

// Scheduler 调度器接口（可插拔）
type Scheduler interface {
	Schedule(tasks []*task_manager.Task) (map[string][]*task_manager.ExecuteTask, error)
	// TransformLogicalMapToExecuteMap 将逻辑任务图转化为物理任务图
	TransformLogicalMapToExecuteMap(logicalMap map[string][]*task_manager.Task) (map[string][]*task_manager.ExecuteTask, error)
	AskForAvailablePorts(logicalMap map[string][]*task_manager.Task) (map[string][]int, error)
}

// UserDefinedScheduler 用户自定义调度器
type UserDefinedScheduler struct {
}

func (u *UserDefinedScheduler) Schedule(tasks []*task_manager.Task) (map[string][]task_manager.ExecuteTask, error) {
	logicalMap := make(map[string][]*task_manager.Task)
	for _, task := range tasks {
		name := task.Locate
		if _, ok := logicalMap[name]; !ok {
			logicalMap[name] = make([]*task_manager.Task, 0)
		}
		logicalMap[name] = append(logicalMap[name], task)
	}
	executeMap, err := u.TransformLogicalMapToExecuteMap(logicalMap)
	if err != nil {
		logger.Errorf("UserDefinedScheduler.TransformLogicalMapToExecuteMap() failed: %v", err)
	}
	return executeMap, nil
}

func (u *UserDefinedScheduler) TransformLogicalMapToExecuteMap(logicalMap map[string][]*task_manager.Task) (map[string][]task_manager.ExecuteTask, error) {
	logger.Fatalf("UserDefinedScheduler.TransformLogicalMapToExecuteMap() not implemented")
	return nil, nil
}

func (u *UserDefinedScheduler) AskForAvailablePorts(logicalMap map[string][]task_manager.Task) (map[string][]int, error) {
	logger.Fatalf("UserDefinedScheduler.AskForAvailablePorts() not implemented")
	return nil, nil
}
