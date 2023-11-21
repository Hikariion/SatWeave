package sun

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/infos"
	task_manager "satweave/shared/task-manager"
	"satweave/utils/errno"
	"satweave/utils/generator"
	"satweave/utils/logger"
	"sync"
	"sync/atomic"
)

// Sun used to help satellite nodes become a group
type Sun struct {
	rpc *messenger.RpcServer
	Server
	leaderInfo                     *infos.NodeInfo
	clusterInfo                    *infos.ClusterInfo
	lastRaftID                     uint64
	mu                             sync.Mutex
	cachedInfo                     map[string]*infos.NodeInfo //cache node info by uuid
	taskRegisteredTaskManagerTable *RegisteredTaskManagerTable
	Scheduler                      *UserDefinedScheduler

	idGenerator *IdGenerator

	checkpointCoordinator *CheckpointCoordinator
}

type IdGenerator struct {
}

type TaskTuple struct {
	TaskManagerId uint64
	ExecuteTask   *common.ExecuteTask
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator{}
}

func (gen *IdGenerator) Next() string {
	uid := uuid.New()
	return uid.String()
}

type Server struct {
	UnimplementedSunServer
}

// MoonRegister give a Raft NodeID to a new edge node
func (s *Sun) MoonRegister(_ context.Context, nodeInfo *infos.NodeInfo) (*RegisterResult, error) {
	hasLeader := true

	// Check Leader info
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.leaderInfo == nil { // This is new leader
		hasLeader = false
		s.leaderInfo = nodeInfo
		s.clusterInfo.LeaderInfo = s.leaderInfo
	}

	if info, ok := s.cachedInfo[nodeInfo.Uuid]; ok {
		return &RegisterResult{
			Result: &common.Result{
				Status: common.Result_OK,
			},
			RaftId:      info.RaftId,
			HasLeader:   hasLeader,
			ClusterInfo: s.clusterInfo,
		}, nil
	}

	// Gen a new Raft NodeID
	raftID := atomic.AddUint64(&s.lastRaftID, 1)
	nodeInfo.RaftId = raftID

	result := RegisterResult{
		Result: &common.Result{
			Status: common.Result_OK,
		},
		RaftId:      raftID,
		HasLeader:   hasLeader,
		ClusterInfo: s.clusterInfo,
	}

	s.cachedInfo[nodeInfo.Uuid] = nodeInfo
	logger.Infof("Register moon success, raftID: %v, leader: %v", raftID, result.ClusterInfo.LeaderInfo.RaftId)
	return &result, nil
}

func (s *Sun) GetLeaderInfo(_ context.Context, nodeInfo *infos.NodeInfo) (*infos.NodeInfo, error) {
	return s.clusterInfo.LeaderInfo, nil
}

func (s *Sun) ReportClusterInfo(_ context.Context, clusterInfo *infos.ClusterInfo) (*common.Result, error) {
	s.mu.Lock()
	s.clusterInfo = clusterInfo
	s.leaderInfo = s.clusterInfo.LeaderInfo
	s.mu.Unlock()
	result := common.Result{
		Status: common.Result_OK,
	}
	return &result, nil
}

func (s *Sun) SubmitJob(ctx context.Context, request *SubmitJobRequest) (*SubmitJobResponse, error) {
	jobId := s.idGenerator.Next()
	executeMap, err := s.innerSubmitJob(ctx, request.Tasks, jobId)
	if err != nil {
		return &SubmitJobResponse{}, status.Errorf(codes.Internal, "submit job failed: %v", err)
	}

	// 把所有 Op 信息注册到 CheckpointCoordinator 里
	err = s.checkpointCoordinator.registerJob(jobId, executeMap)
	if err != nil {
		return &SubmitJobResponse{}, errno.RegisterJobFail
	}

	return &SubmitJobResponse{
		JobId: jobId,
	}, nil
}

func (s *Sun) innerSubmitJob(ctx context.Context, tasks []*common.Task, jobId string) (map[uint64][]*common.ExecuteTask, error) {
	// scheduler
	_, executeMap, err := s.Scheduler.Schedule(0, tasks)
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

func (s *Sun) DeployExecuteTasks(ctx context.Context, jobId string, executeMap map[uint64][]*common.ExecuteTask) error {
	for taskManagerId, executeTasks := range executeMap {
		host := s.taskRegisteredTaskManagerTable.table[taskManagerId].Host
		port := s.taskRegisteredTaskManagerTable.table[taskManagerId].Port
		conn, err := messenger.GetRpcConn(host, port)
		if err != nil {
			logger.Errorf("get rpc conn failed: %v", err)
			return err
		}
		client := task_manager.NewTaskManagerServiceClient(conn)
		// 每个 Execute task 都需要 deploy
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

func (s *Sun) StartExecuteTasks(logicalMap map[uint64][]*common.Task, executeMap map[uint64][]*common.ExecuteTask) error {
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
	for taskManagerId, executeTasks := range executeMap {
		for _, executeTask := range executeTasks {
			subTaskNameInvertedIndex[executeTask.SubtaskName] = &TaskTuple{
				TaskManagerId: taskManagerId,
				ExecuteTask:   executeTask,
			}
		}
	}

	startedTasks := make(map[string]bool)
	for clsName, _ := range taskNameInvertedIndex {
		if _, ok := startedTasks[clsName]; !ok {
			s.dfsToStartExecuteTask(clsName, nextLogicalTasks, taskNameInvertedIndex, subTaskNameInvertedIndex, startedTasks)
		}
	}

	return nil
}

func (s *Sun) dfsToStartExecuteTask(clsName string, nextLogicalTasks map[string][]string, logicalTaskNameInvertedIndex map[string]*common.Task,
	subtaskNameInvertedIndex map[string]*TaskTuple, startedTasks map[string]bool) {
	if _, ok := startedTasks[clsName]; ok {
		return
	}
	startedTasks[clsName] = true
	list, ok := nextLogicalTasks[clsName]
	if ok {
		for _, nextLogicalTaskName := range list {
			if _, exists := startedTasks[nextLogicalTaskName]; !exists {
				s.dfsToStartExecuteTask(nextLogicalTaskName, nextLogicalTasks, logicalTaskNameInvertedIndex, subtaskNameInvertedIndex, startedTasks)
			}
		}
	}

	logicalTask := logicalTaskNameInvertedIndex[clsName]
	for i := 0; i < int(logicalTask.Currency); i++ {
		subtaskName := s.getSubTaskName(clsName, i, int(logicalTask.Currency))
		taskManagerId := subtaskNameInvertedIndex[subtaskName].TaskManagerId
		executeTask := subtaskNameInvertedIndex[subtaskName].ExecuteTask
		s.innerDfsToStartExecuteTask(taskManagerId, executeTask)
	}

}

func (s *Sun) innerDfsToStartExecuteTask(taskManagerID uint64, executeTask *common.ExecuteTask) {
	host := s.taskRegisteredTaskManagerTable.table[taskManagerID].Host
	port := s.taskRegisteredTaskManagerTable.table[taskManagerID].Port
	conn, err := messenger.GetRpcConn(host, port)
	if err != nil {
		logger.Errorf("Fail to get rpc conn on TaskManager %v", taskManagerID)
	}
	client := task_manager.NewTaskManagerServiceClient(conn)
	_, err = client.StartTask(context.Background(), &task_manager.StartTaskRequest{
		SubtaskName: executeTask.SubtaskName,
	})
	if err != nil {
		logger.Errorf("Fail to start subtask: %v on task manager id: ", executeTask.SubtaskName, taskManagerID)
	}
	return
}

func (s *Sun) RegisterTaskManager(_ context.Context, desc *common.TaskManagerDescription) (*common.NilResponse, error) {
	err := s.taskRegisteredTaskManagerTable.register(desc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "register task manager failed: %v", err)
	}
	return &common.NilResponse{}, nil
}

func (s *Sun) GetRegisterTaskManagerTable(context.Context, *common.NilRequest) (*TaskManagerResult, error) {
	return &TaskManagerResult{
		TaskManagerTable: s.taskRegisteredTaskManagerTable.table,
	}, nil
}

func (s *Sun) PrintTaskManagerTable() {
	logger.Infof("Sun printing task manager table...")
	logger.Infof("%v", s.taskRegisteredTaskManagerTable.table)
	logger.Infof("%v", s.Scheduler.RegisteredTaskManagerTable)
}

func (s *Sun) TriggerCheckpoint(_ context.Context, request *TriggerCheckpointRequest) (*TriggerCheckpointResponse, error) {
	checkpointId := generator.GetDataIdGeneratorInstance().Next()
	err := s.checkpointCoordinator.triggerCheckpoint(request.JobId, checkpointId, request.CancelJob)
	if err != nil {
		logger.Errorf("trigger checkpoint failed: %v", err)
		return &TriggerCheckpointResponse{
			Status: &common.Status{
				ErrCode: 1,
				Message: err.Error(),
			},
		}, errno.TriggerCheckpointFail
	}
	return &TriggerCheckpointResponse{
		Status:       &common.Status{},
		CheckpointId: checkpointId,
	}, nil
}

func (s *Sun) getSubTaskName(clsName string, idx, currency int) string {
	return fmt.Sprintf("%s#(%d/%d)", clsName, idx+1, currency)
}

func (s *Sun) AcknowledgeCheckpoint(_ context.Context, request *AcknowledgeCheckpointRequest) (*common.NilResponse, error) {
	if request.Status.ErrCode != 0 {
		logger.Errorf("Failed to acknowledge checkpoint: Status.ErrCode != 0, %s", request.Status.Message)
		return &common.NilResponse{}, errno.AcknowledgeCheckpointFail
	}
	succ, err := s.checkpointCoordinator.AcknowledgeCheckpoint(request)
	if err != nil {
		logger.Errorf("Failed to acknowledge checkpoint: %v", err)
		return &common.NilResponse{}, errno.AcknowledgeCheckpointFail
	}
	if succ {
		logger.Infof("Successfully acknowledge checkpoint(id=%v) of job(id=%v)", request.CheckpointId, request.JobId)
	}
	return &common.NilResponse{}, nil
}

func NewSun(rpc *messenger.RpcServer) *Sun {
	sun := Sun{
		rpc:        rpc,
		leaderInfo: nil,
		clusterInfo: &infos.ClusterInfo{
			LeaderInfo: nil,
			NodesInfo:  nil,
		},
		lastRaftID:                     0,
		mu:                             sync.Mutex{},
		cachedInfo:                     map[string]*infos.NodeInfo{},
		taskRegisteredTaskManagerTable: newRegisteredTaskManagerTable(),
		idGenerator:                    NewIdGenerator(),
	}
	sun.checkpointCoordinator = NewCheckpointCoordinator(sun.taskRegisteredTaskManagerTable)
	sun.Scheduler = newUserDefinedScheduler(sun.taskRegisteredTaskManagerTable)
	RegisterSunServer(rpc, &sun)
	return &sun
}
