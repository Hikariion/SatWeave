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
	// TransformLogicalMapToExecuteMap 将逻辑任务图转化为物理任务图
	TransformLogicalMapToExecuteMap(logicalMap map[uint64][]*common.Task) (map[uint64][]*common.ExecuteTask, error)
	// AskForAvailableWorkers 分配可用的 worker
	AskForAvailableWorkers(logicalMap map[uint64]*[]common.Task) (map[uint64][]int, error)
}

// DefaultScheduler 默认调度器
type DefaultScheduler struct {
	RegisteredTaskManagerTable *RegisteredTaskManagerTable
}

func (u *DefaultScheduler) Schedule(jobId string, tasks []*common.Task) (map[string][]*common.Task, map[string][]*common.ExecuteTask, error) {
	logger.Infof("Begin to scheduler ...")
	logicalMap := make(map[string][]*common.Task) // string node  -> task
	for _, task := range tasks {
		locate := task.Locate
		if _, ok := logicalMap[locate]; !ok {
			logicalMap[locate] = make([]*common.Task, 0)
		}
		logicalMap[locate] = append(logicalMap[locate], task)
	}
	executeMap, err := u.TransformLogicalMapToExecuteMap(jobId, logicalMap, tasks)
	if err != nil {
		logger.Errorf("DefaultScheduler.TransformLogicalMapToExecuteMap() failed: %v", err)
		return nil, nil, err
	}
	return logicalMap, executeMap, nil
}

func (u *DefaultScheduler) TransformLogicalMapToExecuteMap(jobId string, logicalMap map[string][]*common.Task, tasks []*common.Task) (map[string][]*common.ExecuteTask, error) {
	// TODO(qiu): get free slot on each node
	availWorkersMap, err := u.AskForAvailableWorkers(logicalMap)
	logger.Infof("==================Map=======================")
	logger.Infof("%v", availWorkersMap)
	if err != nil {
		logger.Errorf("DefaultScheduler.AskForAvailablePorts() failed: %v", err)
		return nil, err
	}
	executeMap := make(map[string][]*common.ExecuteTask)      // taskmanager_name -> execute_tasks
	nameToExecuteTask := make(map[string]*common.ExecuteTask) // subtask_name -> execute_task
	for satelliteName, logicalTasks := range logicalMap {
		HostPort := u.RegisteredTaskManagerTable.getHostPort(satelliteName)
		executeMap[satelliteName] = make([]*common.ExecuteTask, 0)
		usedWorkerIdx := 0
		for _, logicalTask := range logicalTasks {
			for i := 0; i < int(logicalTask.Currency); i++ {
				workerID := availWorkersMap[satelliteName][usedWorkerIdx]
				usedWorkerIdx++
				logger.Infof("satellite %v availWorkersMap worker id %v", satelliteName, availWorkersMap[satelliteName])
				subTaskName := u.getSubTaskName(jobId, logicalTask.ClsName, i, int(logicalTask.Currency))
				executeTask := &common.ExecuteTask{
					Locate:       satelliteName,
					ClsName:      logicalTask.ClsName,
					SubtaskName:  subTaskName,
					PartitionIdx: int64(i),
					HostPort:     HostPort,
					WorkerId:     workerID,
				}
				executeMap[satelliteName] = append(executeMap[satelliteName], executeTask)
				nameToExecuteTask[subTaskName] = executeTask
			}
		}
	}

	for _, logicalTasks := range logicalMap {
		logger.Infof("logicalTasks: %v", logicalTasks)
		for _, logicalTask := range logicalTasks {

			clsName := logicalTask.ClsName
			if len(logicalTask.InputTasks) > 1 {
				// TODO(qiu)：只支持单个input_task
				logger.Fatalf("Assertion failed: len(logical_task.input_tasks) <= 1")
			}
			for _, inputTaskName := range logicalTask.InputTasks {
				// 找前驱节点，获取并发数
				predecessor := new(common.Task)
				for _, taskP := range tasks {
					logger.Infof("taskP %v clsName %v", taskP.ClsName, inputTaskName)
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
					preSubTaskName := u.getSubTaskName(jobId, predecessor.ClsName, i, int(predecessor.Currency))
					preExecuteTask := nameToExecuteTask[preSubTaskName]
					for j := 0; j < int(logicalTask.Currency); j++ {
						currentSubTaskName := u.getSubTaskName(jobId, logicalTask.ClsName, j, int(logicalTask.Currency))
						currentExecuteTask := nameToExecuteTask[currentSubTaskName]
						logger.Infof("add input_endpoint: %v -> %v", preExecuteTask.SubtaskName, currentExecuteTask.SubtaskName)
						currentExecuteTask.InputEndpoints = append(currentExecuteTask.InputEndpoints,
							&common.InputEndpoints{SatelliteName: preExecuteTask.Locate, HostPort: preExecuteTask.HostPort,
								WorkerId: preExecuteTask.WorkerId})
						preExecuteTask.OutputEndpoints = append(preExecuteTask.OutputEndpoints,
							&common.OutputEndpoints{SatelliteName: currentExecuteTask.Locate, HostPort: currentExecuteTask.HostPort,
								WorkerId: currentExecuteTask.WorkerId})
					}
				}

			}
		}
	}
	return executeMap, nil
}

func (u *DefaultScheduler) AskForAvailableWorkers(logicalMap map[string][]*common.Task) (map[string][]uint64, error) {
	availableWorkersMap := make(map[string][]uint64)
	// TODO： 按需分配 worker
	for satelliteName, tasks := range logicalMap {
		HostPort := u.RegisteredTaskManagerTable.getHostPort(satelliteName)
		host := HostPort.GetHost()
		port := HostPort.GetPort()
		conn, err := messenger.GetRpcConn(host, port)
		if err != nil {
			logger.Errorf("UserDefinedScheduler.AskForAvailablePorts() failed: %v", err)
			return nil, err
		}
		slotNum := 0
		for _, task := range tasks {
			slotNum += int(task.Currency)
		}
		client := task_manager.NewTaskManagerServiceClient(conn)
		resp, err := client.RequestSlot(context.Background(),
			&task_manager.RequiredSlotRequest{RequestSlotNum: uint64(slotNum)})
		if err != nil {
			logger.Errorf("UserDefinedScheduler.AskForAvailablePorts() failed: %v", err)
			return nil, err
		}
		availableWorkersMap[satelliteName] = resp.AvailableWorkers
	}
	return availableWorkersMap, nil
}

func (u *DefaultScheduler) getSubTaskName(jobId string, clsName string, idx, currency int) string {
	return fmt.Sprintf("%s#%s#-%d-%d", jobId, clsName, idx+1, currency)
}

func newUserDefinedScheduler(table *RegisteredTaskManagerTable) *DefaultScheduler {
	return &DefaultScheduler{
		RegisteredTaskManagerTable: table,
	}
}
