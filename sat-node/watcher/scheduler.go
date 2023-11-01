package watcher

import "satweave/sat-node/task-manager"

// Scheduler 调度器接口（可插拔）
type Scheduler interface {
	Schedule(task []task_manager.Task) (map[string][]task_manager.ExecuteTask, error)
	TransformLogicalMapToExecuteMap(logicalMap map[string][]task_manager.ExecuteTask) (map[string][]task_manager.ExecuteTask, error)
	AskForAvailablePorts(logicalMap map[string][]task_manager.Task) (map[string][]int, error)
}

// UserDefinedScheduler 用户自定义调度器
type UserDefinedScheduler struct {
}
