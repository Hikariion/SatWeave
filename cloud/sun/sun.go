package sun

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/infos"
	common2 "satweave/utils/common"
	"satweave/utils/logger"
	"sync"
	"sync/atomic"
)

// Sun used to help satellite nodes become a group
type Sun struct {
	rpc *messenger.RpcServer
	Server
	leaderInfo  *infos.NodeInfo
	clusterInfo *infos.ClusterInfo
	lastRaftID  uint64
	mu          sync.Mutex
	cachedInfo  map[string]*infos.NodeInfo //cache node info by uuid

	StreamHelper *StreamHelper
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

// --------------------------- for stream task --------------------------------
func (s *Sun) RegisterTaskManager(ctx context.Context, request *RegisterTaskManagerRequest) (*RegisterTaskManagerResponse, error) {
	return s.StreamHelper.RegisterTaskManager(ctx, request)
}

func (s *Sun) SaveSnapShot(_ context.Context, request *SaveSnapShotRequest) (*common.Result, error) {
	filePath := request.FilePath
	data := request.State
	err := s.StreamHelper.SaveSnapShot(filePath, data)
	if err != nil {
		return &common.Result{
			Status: common.Result_FAIL,
		}, err
	}
	return &common.Result{
		Status: common.Result_OK,
	}, nil
}

func (s *Sun) RestoreFromCheckpoint(_ context.Context, request *RestoreFromCheckpointRequest) (*RestoreFromCheckpointResponse, error) {
	state, err := s.StreamHelper.RestoreFromCheckpoint(request.SubtaskName)
	if err != nil {
		return &RestoreFromCheckpointResponse{
			Success: false,
			State:   nil,
		}, err
	}
	return &RestoreFromCheckpointResponse{
		Success: true,
		State:   state,
	}, nil
}

func (s *Sun) SubmitJob(ctx context.Context, request *SubmitJobRequest) (*SubmitJobResponse, error) {
	yamlBytes := request.YamlByte
	var tasksWrapper common2.UserTaskWrapper
	err := yaml.Unmarshal(yamlBytes, &tasksWrapper)
	if err != nil {
		logger.Errorf("ReadUserDefinedTasks() failed: %v", err)
		return nil, err
	}
	logicalTasks, err := common2.ConvertUserTaskWrapperToLogicTasks(&tasksWrapper)
	if err != nil {
		logger.Errorf("ConvertUserTaskWrapperToLogicTasks() failed: %v", err)
		return &SubmitJobResponse{
			Success: false,
		}, err
	}

	for _, task := range logicalTasks {
		task.Locate = request.SatelliteName
	}

	logicalTaskMap, executeTaskMap, err := s.StreamHelper.Scheduler.Schedule(request.JobId, logicalTasks)

	err = s.StreamHelper.DeployExecuteTasks(ctx, request.JobId, executeTaskMap)
	if err != nil {
		return &SubmitJobResponse{
			Success: false,
		}, status.Errorf(codes.Internal, "submit job failed: %v", err)
	}

	err = s.StreamHelper.StartExecuteTasks(request.JobId, logicalTaskMap, executeTaskMap)
	if err != nil {
		return &SubmitJobResponse{
			Success: false,
		}, status.Errorf(codes.Internal, "submit job failed: %v", err)
	}

	return &SubmitJobResponse{
		Success: true,
	}, nil
}

// PrintTaskManagerTable For debug
func (s *Sun) PrintTaskManagerTable() {
	s.StreamHelper.PrintTaskManagerTable()
}

func NewSun(rpc *messenger.RpcServer) *Sun {
	sun := Sun{
		rpc:        rpc,
		leaderInfo: nil,
		clusterInfo: &infos.ClusterInfo{
			LeaderInfo: nil,
			NodesInfo:  nil,
		},
		lastRaftID: 0,
		mu:         sync.Mutex{},
		cachedInfo: map[string]*infos.NodeInfo{},

		StreamHelper: NewStreamHelper(),
	}
	RegisterSunServer(rpc, &sun)
	return &sun
}
