package sun

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/infos"
	"satweave/sat-node/watcher"
	common2 "satweave/utils/common"
	"satweave/utils/logger"
	"sync"
)

// Sun used to help satellite nodes become a group
type Sun struct {
	rpc *messenger.RpcServer
	Server
	leaderInfo  *infos.NodeInfo
	clusterInfo *infos.ClusterInfo
	lastRaftID  uint64
	mu          sync.Mutex
	cachedInfo  map[uint64]*infos.NodeInfo //cache node info by uuid

	StreamHelper *StreamHelper

	TaskInfoList *taskInfoList
}

type taskInfoList struct {
	mu       sync.Mutex
	taskInfo []*infos.TaskInfo
}

func (t *taskInfoList) UpdateTaskInfo(taskInfo *infos.TaskInfo) {
	t.mu.Lock()
	t.taskInfo = append(t.taskInfo, taskInfo)
	t.mu.Unlock()
}

func (t *taskInfoList) GetTaskList() []*infos.TaskInfo {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.taskInfo
}

type Server struct {
	UnimplementedSunServer
}

// MoonRegister give a Raft NodeID to a new edge node
func (s *Sun) MoonRegister(_ context.Context, nodeInfo *infos.NodeInfo) (*RegisterResult, error) {
	s.cachedInfo[nodeInfo.RaftId] = nodeInfo

	result := RegisterResult{
		Result: &common.Result{
			Status: common.Result_OK,
		},
	}

	logger.Infof("Register moon success, raftID: %v", nodeInfo.RaftId)
	return &result, nil
}

func (s *Sun) GetLeaderInfo(_ context.Context, nodeInfo *infos.NodeInfo) (*infos.NodeInfo, error) {
	return s.clusterInfo.LeaderInfo, nil
}

func (s *Sun) UpdateTaskInfo(_ context.Context, info *infos.TaskInfo) (*common.Result, error) {
	s.TaskInfoList.UpdateTaskInfo(info)
	return &common.Result{
		Status: common.Result_OK,
	}, nil
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

func (s *Sun) SubmitOfflineJob(request *watcher.GeoUnSensitiveTaskRequest) error {
	accessNodeId := request.AccessNodeId
	info, err := s.GetNodeInfoById(context.Background(), &RaftId{Id: accessNodeId})
	if err != nil {
		return err
	}
	conn, _ := messenger.GetRpcConnByNodeInfo(info)
	client := watcher.NewWatcherClient(conn)
	_, err = client.SubmitGeoUnSensitiveTask(context.Background(), request)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sun) SubmitJob(ctx context.Context, request *SubmitJobRequest) (*SubmitJobResponse, error) {
	yamlBytes := []byte(request.YamlStr)
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

	err = s.StreamHelper.DeployExecuteTasks(ctx, request.JobId, executeTaskMap, request.PathNodes, yamlBytes)
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

func (s *Sun) ReceiverStreamData(_ context.Context, request *ReceiverStreamDataRequest) (*common.Result, error) {
	err := s.StreamHelper.SaveStreamJobData(request.JobId, request.DataId, request.Res)
	if err != nil {
		return &common.Result{
			Status: common.Result_FAIL,
		}, err
	}
	return &common.Result{
		Status: common.Result_OK,
	}, err
}

// PrintTaskManagerTable For debug
func (s *Sun) PrintTaskManagerTable() {
	s.StreamHelper.PrintTaskManagerTable()
}

func (s *Sun) GetNodeInfoById(_ context.Context, raftId *RaftId) (*infos.NodeInfo, error) {
	info, ok := s.cachedInfo[raftId.Id]
	if !ok {
		return nil, nil
	}
	return info, nil
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
		cachedInfo: map[uint64]*infos.NodeInfo{},
		TaskInfoList: &taskInfoList{
			mu: sync.Mutex{},
		},
		StreamHelper: NewStreamHelper(),
	}
	RegisterSunServer(rpc, &sun)
	return &sun
}
