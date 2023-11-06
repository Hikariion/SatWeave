package sun

import (
	"context"
	"fmt"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/shared/task-manager"
	"satweave/utils/logger"
)

// Scheduler 调度器接口（可插拔）
type Scheduler interface {
	// Schedule 调度任务
	// TODO(qiu): clusterID 在做迁移的时候要用到，目前在部署阶段先不用
	Schedule(clusterID int, tasks []*common.Task) (map[string][]*common.ExecuteTask, error)
	// FullFillLogicTask 填充 Task 的 locate
	FullFillLogicTask(tasks []*common.Task) ([]*common.Task, error)
	// TransformLogicalMapToExecuteMap 将逻辑任务图转化为物理任务图
	TransformLogicalMapToExecuteMap(logicalMap map[uint64][]*common.Task) (map[uint64][]*common.ExecuteTask, error)
	// AskForAvailableWorkers 分配可用的 worker TODO 注意，每个并发数都要占用一个 worker
	AskForAvailableWorkers(logicalMap map[uint64]*[]common.Task) (map[uint64][]int, error)
}

// UserDefinedScheduler 用户自定义调度器
// Can define different Task on different TaskManager
type UserDefinedScheduler struct {
	RegisteredTaskManagerTable *RegisteredTaskManagerTable
}

func (u *UserDefinedScheduler) FullFillLogicTask(tasks []*common.Task) ([]*common.Task, error) {
	logger.Fatalf("UserDefinedScheduler.FullFillLogicTask() not implemented")
	return nil, nil
}

func (u *UserDefinedScheduler) Schedule(clusterId int, tasks []*common.Task) (map[uint64][]*common.Task, map[uint64][]*common.ExecuteTask, error) {
	// TODO(qiu): 这里的 string 是节点的标识，之后需要根据 cluster info 里的信息进行修改
	logger.Infof("Begin to scheduler ...")
	logicalMap := make(map[uint64][]*common.Task) // string node  -> task
	for _, task := range tasks {
		locate := task.Locate
		if _, ok := logicalMap[locate]; !ok {
			logicalMap[locate] = make([]*common.Task, 0)
		}
		logicalMap[locate] = append(logicalMap[locate], task)
	}
	executeMap, err := u.TransformLogicalMapToExecuteMap(clusterId, logicalMap)
	if err != nil {
		logger.Errorf("UserDefinedScheduler.TransformLogicalMapToExecuteMap() failed: %v", err)
		return nil, nil, err
	}
	return logicalMap, executeMap, nil
}

func (u *UserDefinedScheduler) TransformLogicalMapToExecuteMap(clusterId int, logicalMap map[uint64][]*common.Task) (map[uint64][]*common.ExecuteTask, error) {
	// TODO(qiu): get free slot on each node
	availWorkersMap, err := u.AskForAvailableWorkers(logicalMap)
	if err != nil {
		logger.Errorf("UserDefinedScheduler.AskForAvailablePorts() failed: %v", err)
		return nil, err
	}
	executeMap := make(map[uint64][]*common.ExecuteTask)      // taskmanager_name -> execute_tasks
	nameToExecuteTask := make(map[string]*common.ExecuteTask) // subtask_name -> execute_task
	for taskManagerID, logicalTasks := range logicalMap {
		taskManagerHost := u.RegisteredTaskManagerTable.getHost(taskManagerID)
		executeMap[taskManagerID] = make([]*common.ExecuteTask, 0)
		usedWorkerIdx := 0
		for _, logicalTask := range logicalTasks {
			for i := 0; i < int(logicalTask.Currency); i++ {
				workerID := availWorkersMap[taskManagerID][usedWorkerIdx]
				usedWorkerIdx++
				subTaskName := u.getSubTaskName(logicalTask.ClsName, i, int(logicalTask.Currency))
				executeTask := &common.ExecuteTask{
					ClsName:      logicalTask.ClsName,
					Resources:    logicalTask.Resources,
					TaskFile:     logicalTask.TaskFile,
					SubtaskName:  subTaskName,
					PartitionIdx: int64(i),
					Host:         taskManagerHost,
					WorkerId:     workerID,
				}
				executeMap[taskManagerID] = append(executeMap[taskManagerID], executeTask)
				nameToExecuteTask[subTaskName] = executeTask
			}
		}
	}

	for _, logicalTasks := range logicalMap {
		for _, logicalTask := range logicalTasks {
			clsName := logicalTask.ClsName
			if len(logicalTask.InputTasks) > 1 {
				// TODO(qiu)：只支持单个input_task
				logger.Fatalf("Assertion failed: len(logical_task.input_tasks) <= 1")
			}
			for _, inputTaskName := range logicalTask.InputTasks {
				// 找前驱节点，获取并发数
				predecessor := new(common.Task)
				for _, taskP := range logicalTasks {
					if taskP.ClsName == inputTaskName {
						predecessor = taskP
						break
					}
				}
				if predecessor == nil {
					logger.Fatalf("Failed: the predecessor task(name=%v) of task(name=%v) is not found", inputTaskName, clsName)
				}
				// 设置 input_endpoints & output_endpoints
				for i := 0; i < int(predecessor.Currency); i++ {
					preSubTaskName := u.getSubTaskName(predecessor.ClsName, i, int(predecessor.Currency))
					preExecuteTask := nameToExecuteTask[preSubTaskName]
					for j := 0; j < int(logicalTask.Currency); j++ {
						currentSubTaskName := u.getSubTaskName(logicalTask.ClsName, j, int(logicalTask.Currency))
						currentExecuteTask := nameToExecuteTask[currentSubTaskName]
						// TODO(qiu): 这里只映射了节点名，Ip 和 端口 需要从 clusterInfo 中获取
						currentExecuteTask.InputEndpoints = append(currentExecuteTask.InputEndpoints,
							&common.InputEndpoints{Host: preExecuteTask.Host, WorkerId: preExecuteTask.WorkerId})
						preExecuteTask.OutputEndpoints = append(preExecuteTask.OutputEndpoints,
							&common.OutputEndpoints{Host: currentExecuteTask.Host, WorkerId: currentExecuteTask.WorkerId})
					}
				}

			}
		}
	}
	return executeMap, nil
}

func (u *UserDefinedScheduler) AskForAvailableWorkers(logicalMap map[uint64][]*common.Task) (map[uint64][]uint64, error) {
	availableWorkersMap := make(map[uint64][]uint64)
	// TODO： 按需分配 worker
	for taskManagerId, tasks := range logicalMap {
		host := u.RegisteredTaskManagerTable.getHost(taskManagerId)
		port := u.RegisteredTaskManagerTable.getPort(taskManagerId)
		conn, err := messenger.GetRpcConn(host, port)
		if err != nil {
			logger.Errorf("UserDefinedScheduler.AskForAvailablePorts() failed: %v", err)
			return nil, err
		}
		client := task_manager.NewTaskManagerServiceClient(conn)
		requiredSlotNum := 0
		for _, task := range tasks {
			requiredSlotNum += int(task.Currency)
		}
		resp, err := client.RequestSlot(context.Background(),
			&task_manager.RequiredSlotRequest{RequestSlotNum: uint64(requiredSlotNum)})
		if err != nil {
			logger.Errorf("UserDefinedScheduler.AskForAvailablePorts() failed: %v", err)
			return nil, err
		}
		availableWorkersMap[taskManagerId] = resp.AvailableWorkers
	}
	return availableWorkersMap, nil
}

func (u *UserDefinedScheduler) getSubTaskName(clsName string, idx, currency int) string {
	return fmt.Sprintf("%s#(%d/%d)", clsName, idx, currency)
}

func newUserDefinedScheduler(table *RegisteredTaskManagerTable) *UserDefinedScheduler {
	return &UserDefinedScheduler{
		RegisteredTaskManagerTable: table,
	}
}
