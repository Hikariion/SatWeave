package sun

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"path"
	"satweave/messenger"
	"satweave/messenger/common"
	task_manager "satweave/shared/task-manager"
	common2 "satweave/utils/common"
	"satweave/utils/generator"
	"satweave/utils/logger"
)

type StreamHelper struct {
	taskRegisteredTaskManagerTable *RegisteredTaskManagerTable
	Scheduler                      *DefaultScheduler
	snapshotDir                    string

	//checkpointCoordinator *CheckpointCoordinator
	//jobInfoDir  string
	//snapshotDir string
}

type TaskTuple struct {
	SatelliteName string
	ExecuteTask   *common.ExecuteTask
}

func (s *StreamHelper) SubmitJob(ctx context.Context, request *SubmitJobRequest) (*SubmitJobResponse, error) {
	jobId := generator.GetJobIdGeneratorInstance().Next()

	_, err := s.innerSubmitJob(ctx, request.Tasks, jobId)

	if err != nil {
		return &SubmitJobResponse{
			JobId:   jobId,
			Success: false,
		}, err
	}

	return &SubmitJobResponse{
		JobId:   jobId,
		Success: true,
	}, nil
}

func (s *StreamHelper) innerSubmitJob(ctx context.Context, tasks []*common.Task, jobId string) (map[string][]*common.ExecuteTask, error) {
	// scheduler
	_, executeMap, err := s.Scheduler.Schedule(jobId, tasks)
	if err != nil {
		logger.Errorf("schedule failed: %v", err)
		return nil, err
	}

	// deploy 创建对应的 worker
	err = s.DeployExecuteTasks(ctx, jobId, executeMap)
	if err != nil {
		logger.Errorf("deploy execute tasks failed: %v", err)
		return nil, err
	}
	logger.Infof("deploy execute tasks success")

	// TODO start 让 worker run起来

	return executeMap, nil
}

func (s *StreamHelper) DeployExecuteTasks(ctx context.Context, jobId string, executeMap map[string][]*common.ExecuteTask) error {
	for taskManagerId, executeTasks := range executeMap {
		HostPort := s.taskRegisteredTaskManagerTable.table[taskManagerId].HostPort
		host := HostPort.GetHost()
		port := HostPort.GetPort()
		conn, err := messenger.GetRpcConn(host, port)
		if err != nil {
			logger.Errorf("get rpc conn failed: %v", err)
			return err
		}
		client := task_manager.NewTaskManagerServiceClient(conn)
		// 每个 Execute task 都需要 deploy
		// deploy的过程其实是创建一个worker的过程
		for _, executeTask := range executeTasks {
			_, err := client.DeployTask(ctx, &task_manager.DeployTaskRequest{
				ExecTask: executeTask,
				JobId:    jobId,
			})
			if err != nil {
				logger.Errorf("deploy task on task manager id: %v, failed: %v", taskManagerId, err)
				return err
			}
		}
		logger.Infof("Deploy all execute task on task manager id: %v, success", taskManagerId)
	}

	return nil
}

func (s *StreamHelper) StartExecuteTasks(jobId string, logicalMap map[string][]*common.Task, executeMap map[string][]*common.ExecuteTask) error {
	// clsName -> Task
	taskNameInvertedIndex := make(map[string]*common.Task)
	for _, tasks := range logicalMap {
		for _, task := range tasks {
			if _, ok := taskNameInvertedIndex[task.ClsName]; !ok {
				taskNameInvertedIndex[task.ClsName] = task
			}
		}
	}

	// clsName -> output task: List[str]
	nextLogicalTasks := make(map[string][]string)
	for _, tasks := range logicalMap {
		for _, task := range tasks {
			for _, preTask := range task.InputTasks {
				if _, ok := nextLogicalTasks[preTask]; !ok {
					nextLogicalTasks[preTask] = make([]string, 0)
				}
				nextLogicalTasks[preTask] = append(nextLogicalTasks[preTask], task.ClsName)
			}
		}
	}

	// subtaskName -> (taskManagerId, ExecuteTask)
	subTaskNameInvertedIndex := make(map[string]*TaskTuple)
	for satelliteName, executeTasks := range executeMap {
		for _, executeTask := range executeTasks {
			subTaskNameInvertedIndex[executeTask.SubtaskName] = &TaskTuple{
				SatelliteName: satelliteName,
				ExecuteTask:   executeTask,
			}
		}
	}

	startedTasks := make(map[string]bool)
	for clsName, _ := range taskNameInvertedIndex {
		if _, ok := startedTasks[clsName]; !ok {
			s.dfsToStartExecuteTask(jobId, clsName, nextLogicalTasks, taskNameInvertedIndex, subTaskNameInvertedIndex, startedTasks)
		}
	}

	return nil
}

func (s *StreamHelper) dfsToStartExecuteTask(jobId string, clsName string, nextLogicalTasks map[string][]string, logicalTaskNameInvertedIndex map[string]*common.Task,
	subtaskNameInvertedIndex map[string]*TaskTuple, startedTasks map[string]bool) {
	if _, ok := startedTasks[clsName]; ok {
		return
	}
	startedTasks[clsName] = true
	list, ok := nextLogicalTasks[clsName]
	if ok {
		for _, nextLogicalTaskName := range list {
			if _, exists := startedTasks[nextLogicalTaskName]; !exists {
				s.dfsToStartExecuteTask(jobId, nextLogicalTaskName, nextLogicalTasks, logicalTaskNameInvertedIndex, subtaskNameInvertedIndex, startedTasks)
			}
		}
	}

	logicalTask := logicalTaskNameInvertedIndex[clsName]
	for i := 0; i < int(logicalTask.Currency); i++ {
		subtaskName := s.getSubTaskName(jobId, clsName, i, int(logicalTask.Currency))
		SatelliteName := subtaskNameInvertedIndex[subtaskName].SatelliteName
		executeTask := subtaskNameInvertedIndex[subtaskName].ExecuteTask
		s.innerDfsToStartExecuteTask(SatelliteName, executeTask)
	}

}

func (s *StreamHelper) innerDfsToStartExecuteTask(satelliteName string, executeTask *common.ExecuteTask) {
	HostPort := s.taskRegisteredTaskManagerTable.table[satelliteName].HostPort
	host := HostPort.GetHost()
	port := HostPort.GetPort()
	conn, err := messenger.GetRpcConn(host, port)
	if err != nil {
		logger.Errorf("Fail to get rpc conn on TaskManager %v", satelliteName)
	}
	client := task_manager.NewTaskManagerServiceClient(conn)
	_, err = client.StartTask(context.Background(), &task_manager.StartTaskRequest{
		SubtaskName: executeTask.SubtaskName,
	})
	if err != nil {
		logger.Errorf("Fail to start subtask: %v on task manager id: ", executeTask.SubtaskName, satelliteName)
	}
	return
}

func (s *StreamHelper) RegisterTaskManager(_ context.Context, request *RegisterTaskManagerRequest) (*RegisterTaskManagerResponse, error) {
	err := s.taskRegisteredTaskManagerTable.register(request.TaskManagerDesc)
	if err != nil {
		return &RegisterTaskManagerResponse{
			Success: false,
		}, status.Errorf(codes.Internal, "register task manager failed: %v", err)
	}
	return &RegisterTaskManagerResponse{
		Success: true,
	}, nil
}

//func (s *StreamHelper) GetRegisterTaskManagerTable(context.Context, *common.NilRequest) (*TaskManagerResult, error) {
//	return &TaskManagerResult{
//		TaskManagerTable: s.taskRegisteredTaskManagerTable.table,
//	}, nil
//}

func (s *StreamHelper) PrintTaskManagerTable() {
	logger.Infof("Sun printing task manager table...")
	logger.Infof("%v", s.taskRegisteredTaskManagerTable.table)
	logger.Infof("%v", s.Scheduler.RegisteredTaskManagerTable)
}

func (s *StreamHelper) getSubTaskName(jobId string, clsName string, idx, currency int) string {
	return fmt.Sprintf("%s#%s#-%d-%d", jobId, clsName, idx+1, currency)
}

func (s *StreamHelper) SaveSnapShot(filePath string, state []byte) error {
	file, err := os.Create(path.Join(s.snapshotDir, filePath))
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = file.Write(state)

	file.Sync()

	return nil
}

func NewStreamHelper() *StreamHelper {
	streamHelper := &StreamHelper{
		taskRegisteredTaskManagerTable: newRegisteredTaskManagerTable(),
		snapshotDir:                    "./snapshot",
	}

	// 创建 snapshot 目录
	_ = common2.InitPath(streamHelper.snapshotDir)

	streamHelper.Scheduler = newUserDefinedScheduler(streamHelper.taskRegisteredTaskManagerTable)
	return streamHelper
}
