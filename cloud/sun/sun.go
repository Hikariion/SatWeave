package sun

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/infos"
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

func (s *Sun) RegisterTaskManager(_ context.Context, request *RegisterTaskManagerRequest) (*common.NilResponse, error) {
	err := s.taskRegisteredTaskManagerTable.register(request.TaskManagerDesc)
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
	}
	RegisterSunServer(rpc, &sun)
	return &sun
}
