package sun

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/infos"
	task_manager "satweave/shared/task-manager"
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
}

type IdGenerator struct {
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
	err := s.innerSubmitJob(ctx, request.Tasks)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "submit job failed: %v", err)
	}
	return &SubmitJobResponse{
		JobId: jobId,
	}, nil
}

func (s *Sun) innerSubmitJob(ctx context.Context, tasks []*common.Task) error {
	// scheduler
	_, executeMap, err := s.Scheduler.Schedule(0, tasks)
	if err != nil {
		logger.Errorf("schedule failed: %v", err)
		return err
	}

	// deploy 创建对应的 worker
	err = s.deployExecuteTasks(ctx, executeMap)
	if err != nil {
		logger.Errorf("deploy execute tasks failed: %v", err)
		return err
	}
	logger.Infof("deploy execute tasks success")

	// start 让 worker run起来

	return nil
}

func (s *Sun) deployExecuteTasks(ctx context.Context, executeMap map[uint64][]*common.ExecuteTask) error {
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
			_, err := client.DeployTask(ctx, executeTask)
			if err != nil {
				logger.Errorf("deploy task on task manager id: %v, failed: %v", taskManagerId, err)
				return err
			}
		}
		logger.Infof("Deploy all execute task on task manager id: %v, success", taskManagerId)
	}

	return nil
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
	sun.Scheduler = newUserDefinedScheduler(sun.taskRegisteredTaskManagerTable)
	RegisterSunServer(rpc, &sun)
	return &sun
}
