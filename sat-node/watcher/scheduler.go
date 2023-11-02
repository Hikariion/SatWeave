package watcher

import (
	"fmt"
	"satweave/sat-node/task-manager"
	"satweave/utils/logger"
)

// Scheduler 调度器接口（可插拔）
type Scheduler interface {
	Schedule(tasks []*task_manager.Task) (map[string][]*task_manager.ExecuteTask, error)
	// FullFillLogicTask 填充 Task 的 locate
	FullFillLogicTask(tasks []*task_manager.Task) ([]*task_manager.Task, error)
	// TransformLogicalMapToExecuteMap 将逻辑任务图转化为物理任务图
	TransformLogicalMapToExecuteMap(logicalMap map[string][]*task_manager.Task) (map[string][]*task_manager.ExecuteTask, error)
}

// UserDefinedScheduler 用户自定义调度器
// Can define different Task on different TaskManager
type UserDefinedScheduler struct {
}

func (u *UserDefinedScheduler) FullFillLogicTask(tasks []*task_manager.Task) ([]*task_manager.Task, error) {
	logger.Fatalf("UserDefinedScheduler.FullFillLogicTask() not implemented")
	return nil, nil
}

func (u *UserDefinedScheduler) Schedule(tasks []*task_manager.Task) (map[string][]*task_manager.ExecuteTask, error) {
	logicalMap := make(map[string][]*task_manager.Task)
	for _, task := range tasks {
		locate := task.Locate
		if _, ok := logicalMap[locate]; !ok {
			logicalMap[locate] = make([]*task_manager.Task, 0)
		}
		logicalMap[locate] = append(logicalMap[locate], task)
	}
	executeMap, err := u.TransformLogicalMapToExecuteMap(logicalMap)
	if err != nil {
		logger.Errorf("UserDefinedScheduler.TransformLogicalMapToExecuteMap() failed: %v", err)
	}
	return executeMap, nil
}

func (u *UserDefinedScheduler) TransformLogicalMapToExecuteMap(logicalMap map[string][]*task_manager.Task) (map[string][]*task_manager.ExecuteTask, error) {
	executeMap := make(map[string][]*task_manager.ExecuteTask)      // taskmanager_name -> execute_tasks
	nameToExecuteTask := make(map[string]*task_manager.ExecuteTask) // subtask_name -> execute_task
	for taskManagerName, logicalTasks := range logicalMap {
		executeMap[taskManagerName] = make([]*task_manager.ExecuteTask, 0)
		for _, logicalTask := range logicalTasks {
			for i := 0; i < int(logicalTask.Currency); i++ {
				subTaskName := u.getSubTaskName(logicalTask.ClsName, i, int(logicalTask.Currency))
				executeTask := &task_manager.ExecuteTask{
					ClsName:      logicalTask.ClsName,
					Resources:    logicalTask.Resources,
					TaskFile:     logicalTask.TaskFile,
					SubtaskName:  subTaskName,
					PartitionIdx: int64(i),
				}
				executeMap[taskManagerName] = append(executeMap[taskManagerName], executeTask)
				nameToExecuteTask[subTaskName] = executeTask
			}
		}
	}

	for _, logicalTasks := range logicalMap {
		for _, logicalTask := range logicalTasks {
			// TODO(qiu)
			//clsName := logicalTask.ClsName
			if len(logicalTask.InputTasks) <= 1 {
				logger.Fatalf("Assertion failed: len(logical_task.input_tasks) == 1")
			}
			for _, inputTaskName := range logicalTask.InputTasks {
				// 找前驱节点，获取并发数
				predecessor := new(task_manager.Task)
				for _, taskP := range logicalTasks {
					if taskP.ClsName == inputTaskName {
						predecessor = taskP
						break
					}
				}
				if predecessor == nil {
					//logger.Fatalf("Failed: the predecessor task(name=%v) of task(name=%v) is not found")
					continue
				}
				// 设置 input_endpoints & output_endpoints
				for i := 0; i < int(predecessor.Currency); i++ {
					preSubTaskName := u.getSubTaskName(predecessor.ClsName, i, int(predecessor.Currency))
					preExecuteTask := nameToExecuteTask[preSubTaskName]
					for j := 0; j < int(logicalTask.Currency); j++ {
						currentSubTaskName := u.getSubTaskName(logicalTask.ClsName, j, int(logicalTask.Currency))
						currentExecuteTask := nameToExecuteTask[currentSubTaskName]
						// TODO(qiu): 这里只映射了节点名，Ip 和 端口 需要从 clusterInfo 中获取
						currentExecuteTask.InputEndpoints = append(currentExecuteTask.InputEndpoints, preExecuteTask.Locate)
						preExecuteTask.OutputEndpoints = append(preExecuteTask.OutputEndpoints, currentExecuteTask.Locate)
					}
				}

			}
		}
	}
	return executeMap, nil
}

func (u *UserDefinedScheduler) AskForAvailablePorts(logicalMap map[string][]task_manager.Task) (map[string][]int, error) {
	logger.Fatalf("UserDefinedScheduler.AskForAvailablePorts() not implemented")
	return nil, nil
}

func (u *UserDefinedScheduler) getSubTaskName(clsName string, idx, currency int) string {
	return fmt.Sprintf("%s#(%d/%d)", clsName, idx, currency)
}
