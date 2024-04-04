package sun

import (
	"context"
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

//func (s *Sun) SubmitJob(ctx context.Context, request *SubmitJobRequest) (*SubmitJobResponse, error) {
//	jobId := s.idGenerator.Next()
//
//	// 创建存储 jobInfo 的目录
//	jobInfoPath := path.Join(s.jobInfoDir, jobId)
//	err := os.MkdirAll(jobInfoPath, 755)
//	if err != nil {
//		logger.Errorf("mkdir job info path failed: %v", err)
//		return &SubmitJobResponse{}, err
//	}
//
//	executeMap, err := s.innerSubmitJob(ctx, request.Tasks, jobId)
//	if err != nil {
//		return &SubmitJobResponse{}, status.Errorf(codes.Internal, "submit job failed: %v", err)
//	}
//
//	// 把所有 Op 信息注册到 CheckpointCoordinator 里
//	err = s.checkpointCoordinator.registerJob(jobId, executeMap)
//	if err != nil {
//		return &SubmitJobResponse{}, errno.RegisterJobFail
//	}
//
//	return &SubmitJobResponse{
//		JobId: jobId,
//	}, nil
//}

//func (s *Sun) TriggerCheckpoint(_ context.Context, request *TriggerCheckpointRequest) (*TriggerCheckpointResponse, error) {
//	checkpointId := generator.GetDataIdGeneratorInstance().Next()
//	err := s.checkpointCoordinator.triggerCheckpoint(request.JobId, checkpointId, request.CancelJob)
//	if err != nil {
//		logger.Errorf("trigger checkpoint failed: %v", err)
//		return &TriggerCheckpointResponse{
//			Status: &common.Status{
//				ErrCode: 1,
//				Message: err.Error(),
//			},
//		}, errno.TriggerCheckpointFail
//	}
//	return &TriggerCheckpointResponse{
//		Status:       &common.Status{},
//		CheckpointId: checkpointId,
//	}, nil
//}

//func (s *Sun) RestoreFromCheckpoint(_ context.Context, request *RestoreFromCheckpointRequest) (*RestoreFromCheckpointResponse, error) {
//	return nil, status.Errorf(codes.Unimplemented, "method RestoreFromCheckpoint not implemented")
//}
//
//func (s *Sun) AcknowledgeCheckpoint(_ context.Context, request *AcknowledgeCheckpointRequest) (*common.NilResponse, error) {
//	if request.Status.ErrCode != 0 {
//		logger.Errorf("Failed to acknowledge checkpoint: Status.ErrCode != 0, %s", request.Status.Message)
//		return &common.NilResponse{}, errno.AcknowledgeCheckpointFail
//	}
//	succ, err := s.checkpointCoordinator.AcknowledgeCheckpoint(request)
//	if err != nil {
//		logger.Errorf("Failed to acknowledge checkpoint: %v", err)
//		return &common.NilResponse{}, errno.AcknowledgeCheckpointFail
//	}
//	if succ {
//		logger.Infof("Successfully acknowledge checkpoint(id=%v) of job(id=%v), begin to save", request.CheckpointId, request.JobId)
//	}
//	// 存储 state
//	stateData := request.State.Content
//	clsName := strings.Split(request.SubtaskName, "#")[0]
//	jobId := request.JobId
//	partitionIdx := strings.Split(request.SubtaskName, "#")[1]
//	partitionIdx = strings.Trim(partitionIdx, "()")
//	partitionIdx = strings.Split(partitionIdx, "/")[0]
//
//	saveFilePath := fmt.Sprintf(s.snapshotDir, jobId, clsName, partitionIdx)
//	err = common2.SaveBytesToFile(stateData, saveFilePath)
//	if err != nil {
//		logger.Errorf("Failed to save state: %v", err)
//		return &common.NilResponse{}, err
//	}
//	return &common.NilResponse{}, nil
//}

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
