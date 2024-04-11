package watcher

import (
	"context"
	"github.com/mohae/deepcopy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/infos"
	"satweave/sat-node/moon"
	moon2 "satweave/shared/moon"
	"satweave/utils/errno"
	"satweave/utils/logger"
	"satweave/utils/timestamp"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Watcher struct {
	ctx          context.Context
	selfNodeInfo *infos.NodeInfo
	moon         moon.InfoController
	Monitor      Monitor
	register     *infos.StorageRegister
	timer        *time.Timer
	timerMutex   sync.Mutex
	Config       *Config

	addNodeMutex sync.Mutex

	currentClusterInfo infos.ClusterInfo
	clusterInfoMutex   sync.RWMutex

	cancelFunc context.CancelFunc

	UnimplementedWatcherServer
}

// AddNewNodeToCluster will propose a new NodeInfo in moon,
// if success, it will propose a ConfChang, to add the raftNode into moon group
func (w *Watcher) AddNewNodeToCluster(_ context.Context, info *infos.NodeInfo) (*AddNodeReply, error) {
	logger.Infof("receive add new node to cluster: %v", info.RaftId)
	w.addNodeMutex.Lock()
	defer w.addNodeMutex.Unlock()

	flag := true // 判断该请求是否合法
	logger.Infof("add new node to cluster: %v", info.RaftId)

	currentPeerInfos := w.GetCurrentPeerInfo()
	for _, peerInfo := range currentPeerInfos {
		if peerInfo.RaftId == info.RaftId {
			flag = false
			w.moon.NodeInfoChanged(info)
			break
		}
	}

	if flag {
		logger.Infof("check ok, start propose node info: %v", info.RaftId)
		request := &moon2.ProposeInfoRequest{
			Head: &common.Head{
				Timestamp: timestamp.Now(),
				Term:      w.GetCurrentTerm(),
			},
			Operate:  moon2.ProposeInfoRequest_ADD,
			Id:       strconv.FormatUint(info.RaftId, 10),
			BaseInfo: &infos.BaseInfo{Info: &infos.BaseInfo_NodeInfo{NodeInfo: info}},
		}
		_, err := w.moon.ProposeInfo(w.ctx, request)
		logger.Infof("propose New nodeInfo: %v", info.RaftId)
		if err != nil {
			// TODO
			return nil, err
		}
		err = w.moon.ProposeConfChangeAddNode(w.ctx, info)
		logger.Infof("propose conf change to add node: %v", info.RaftId)
		if err != nil {
			// TODO
			return nil, err
		}
	}

	return &AddNodeReply{
		Result: &common.Result{
			Status: common.Result_OK,
		},
		PeersNodeInfo: w.GetCurrentPeerInfo(),
		LeaderInfo:    nil,
	}, nil
}

// GetClusterInfo return requested cluster info to rpc client,
// if GetClusterInfoRequest.Term == 0, it will return current cluster info.
func (w *Watcher) GetClusterInfo(_ context.Context, request *GetClusterInfoRequest) (*GetClusterInfoReply, error) {
	info, err := w.GetClusterInfoByTerm(request.Term)
	if err != nil {
		return &GetClusterInfoReply{
			Result: &common.Result{
				Status:  common.Result_FAIL,
				Message: err.Error(),
			},
			ClusterInfo: nil,
		}, err
	}
	return &GetClusterInfoReply{
		Result: &common.Result{
			Status: common.Result_OK,
		},
		ClusterInfo: &info,
	}, nil
}

// GetClusterInfoByTerm return cluster info directly, it is called by other components
// on same node.
func (w *Watcher) GetClusterInfoByTerm(term uint64) (infos.ClusterInfo, error) {
	if term == 0 {
		info := w.GetCurrentClusterInfo()
		if info.Term == 0 {
			return info, errno.InfoNotFound
		}
		return info, nil
	}
	clusterInfoStorage := w.register.GetStorage(infos.InfoType_CLUSTER_INFO)
	info, err := clusterInfoStorage.Get(strconv.FormatUint(term, 10))
	if err != nil {
		return infos.ClusterInfo{}, err
	}
	return *info.BaseInfo().GetClusterInfo(), nil
}

func (w *Watcher) GetCurrentClusterInfo() infos.ClusterInfo {
	w.clusterInfoMutex.RLock()
	defer w.clusterInfoMutex.RUnlock()
	return w.currentClusterInfo
}

func (w *Watcher) SetOnInfoUpdate(infoType infos.InfoType, name string, f infos.StorageUpdateFunc) error {
	storage := w.register.GetStorage(infoType)
	storage.SetOnUpdate(name, f)
	return nil
}

func (w *Watcher) GetCurrentTerm() uint64 {
	return w.GetCurrentClusterInfo().Term
}

// 获取所有对等点的信息，该信息保证通信节点真实有效
func (w *Watcher) GetCurrentPeerInfo() []*infos.NodeInfo {
	nodeInfoStorage := w.register.GetStorage(infos.InfoType_NODE_INFO)
	nodeInfos, err := nodeInfoStorage.GetAll()
	if err != nil {
		logger.Errorf("get nodeInfo from nodeInfoStorage fail: %v", err)
		return nil
	}
	var peerNodes []*infos.NodeInfo
	for _, info := range nodeInfos {
		peerNodes = append(peerNodes, info.BaseInfo().GetNodeInfo())
	}
	sort.Slice(peerNodes, func(i, j int) bool {
		return peerNodes[i].RaftId > peerNodes[j].RaftId
	})
	return peerNodes
}

func (w *Watcher) genNewMapXClusterInfo() *infos.ClusterInfo {
	oldClusterInfo := w.GetCurrentClusterInfo()
	nodeInfoStorage := w.register.GetStorage(infos.InfoType_NODE_INFO)
	nodeInfos, err := nodeInfoStorage.GetAll()
	if err != nil {
		logger.Errorf("get nodeInfo from nodeInfoStorage fail: %v", err)
		return nil
	}

	oldNodesId := make(map[uint64]struct{})
	for _, node := range oldClusterInfo.NodesInfo {
		oldNodesId[node.RaftId] = struct{}{}
	}

	var clusterNodes []*infos.NodeInfo
	for _, info := range nodeInfos {
		// copy before change to avoid data race
		nodeInfo := deepcopy.Copy(info.BaseInfo().GetNodeInfo()).(*infos.NodeInfo)
		report := w.Monitor.GetNodeReport(nodeInfo.RaftId)
		if report == nil {
			logger.Warningf("get report: %v from Monitor fail", nodeInfo.RaftId)
			logger.Warningf("set node: %v state OFFLINE", nodeInfo.RaftId)
			nodeInfo.State = infos.NodeState_OFFLINE
		} else {
			nodeInfo.State = report.State
		}
		// check if already in oldNodes
		if len(nodeInfos) > len(oldNodesId) {
			if _, ok := oldNodesId[nodeInfo.RaftId]; ok {
				nodeInfo.Capacity = 0
			}
		}
		clusterNodes = append(clusterNodes, nodeInfo)
	}

	sort.Slice(clusterNodes, func(i, j int) bool {
		return clusterNodes[i].RaftId < clusterNodes[j].RaftId
	})
	leaderID := w.moon.GetLeaderID()
	// TODO : need a way to gen infoStorage key
	leaderInfo, err := nodeInfoStorage.Get(strconv.FormatUint(leaderID, 10))
	if err != nil {
		logger.Errorf("get leaderInfo from nodeInfoStorage fail: %v", err)
		return nil
	}
	return &infos.ClusterInfo{
		Term:            uint64(time.Now().UnixNano()),
		LeaderInfo:      leaderInfo.BaseInfo().GetNodeInfo(),
		NodesInfo:       clusterNodes,
		UpdateTimestamp: nil,
		LastTerm:        w.GetCurrentTerm(),
	}
}

func (w *Watcher) genNewClusterInfo() *infos.ClusterInfo {
	nodeInfoStorage := w.register.GetStorage(infos.InfoType_NODE_INFO)
	nodeInfos, err := nodeInfoStorage.GetAll()
	if err != nil {
		logger.Errorf("get nodeInfo from nodeInfoStorage fail: %v", err)
		return nil
	}
	var clusterNodes []*infos.NodeInfo
	for _, info := range nodeInfos {
		// copy before change to avoid data race
		nodeInfo := deepcopy.Copy(info.BaseInfo().GetNodeInfo()).(*infos.NodeInfo)
		report := w.Monitor.GetNodeReport(nodeInfo.RaftId)
		if report == nil {
			logger.Warningf("get report: %v from Monitor fail", nodeInfo.RaftId)
			logger.Warningf("set node: %v state OFFLINE", nodeInfo.RaftId)
			nodeInfo.State = infos.NodeState_OFFLINE
		} else {
			nodeInfo.State = report.State
		}
		clusterNodes = append(clusterNodes, nodeInfo)
	}
	sort.Slice(clusterNodes, func(i, j int) bool {
		return clusterNodes[i].RaftId < clusterNodes[j].RaftId
	})
	leaderID := w.moon.GetLeaderID()
	// TODO : need a way to gen infoStorage key
	leaderInfo, err := nodeInfoStorage.Get(strconv.FormatUint(leaderID, 10))
	if err != nil {
		logger.Errorf("get leaderInfo from nodeInfoStorage fail: %v", err)
		return nil
	}
	return &infos.ClusterInfo{
		Term:            uint64(time.Now().UnixNano()),
		LeaderInfo:      leaderInfo.BaseInfo().GetNodeInfo(),
		NodesInfo:       clusterNodes,
		UpdateTimestamp: nil,
		LastTerm:        w.GetCurrentTerm(),
	}
}

func (w *Watcher) RequestJoinCluster(leaderInfo *infos.NodeInfo) error {
	if leaderInfo == nil || leaderInfo.RaftId == w.selfNodeInfo.RaftId {
		// 此节点为集群中第一个节点
		w.moon.Set(w.selfNodeInfo, leaderInfo, nil)
		return nil
	}
	conn, err := messenger.GetRpcConnByNodeInfo(leaderInfo)
	if err != nil {
		logger.Errorf("Request Join group err: %v", err.Error())
		return err
	}
	client := NewWatcherClient(conn)
	logger.Infof("send Join group request to leader")
	reply, err := client.AddNewNodeToCluster(w.ctx, w.selfNodeInfo)
	if err != nil {
		logger.Errorf("Request join to group err: %v", err.Error())
		return err
	}
	w.moon.Set(w.selfNodeInfo, leaderInfo, reply.PeersNodeInfo)
	return nil
}

func (w *Watcher) StartMoon() {
	go w.moon.Run()
}

func (w *Watcher) GetMoon() moon.InfoController {
	return w.moon
}

func (w *Watcher) AskSky() (leaderInfo *infos.NodeInfo, err error) {
	// TODO(qiu): Init RaftId by satellite name
	if w.Config.SunAddr == "" {
		logger.Errorf("sun addr is empty")
		return nil, errno.ConnectSunFail
	}
	conn, err := grpc.Dial(w.Config.SunAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := sun.NewSunClient(conn)
	result, err := c.MoonRegister(context.Background(), w.selfNodeInfo)
	if err != nil {
		return nil, err
	}
	// update raftID for watcher and moon
	w.selfNodeInfo.RaftId = result.RaftId
	return result.ClusterInfo.LeaderInfo, nil
}

func (w *Watcher) GetSelfInfo() *infos.NodeInfo {
	return w.selfNodeInfo
}

func (w *Watcher) Run() {
	// start Monitor
	go w.Monitor.Run()
	// watch Monitor
	go w.processMonitor()
	leaderInfo, err := w.AskSky()
	if err != nil {
		logger.Warningf("watcher ask sky err: %v", err)
	} else {
		logger.Infof("watcher ask sky success, leader: %v, addr: %v", leaderInfo.RaftId, leaderInfo.IpAddr)
		err = w.RequestJoinCluster(leaderInfo)
		if err != nil {
			logger.Errorf("watcher request join to cluster err: %v", err)
		}
	}
	logger.Infof("%v request join to cluster success", w.GetSelfInfo().RaftId)
	w.StartMoon()
	logger.Infof("moon init success, NodeID: %v", w.GetSelfInfo().RaftId)
}

func (w *Watcher) Stop() {
	w.Monitor.Stop()
	w.moon.Stop()
	w.cancelFunc()
}

func (w *Watcher) processMonitor() {
	c := w.Monitor.GetEventChannel()
	for {
		select {
		case <-w.ctx.Done():
			return
		case reportEvent := <-c:
			logger.Warningf("watcher receive Monitor event")
			// Process Monitor event
			w.nodeInfoChanged(nil)
			nodeId := reportEvent.Report.NodeId
			if reportEvent.Report.State == infos.NodeState_OFFLINE {
				// Migrate Task
				onlineNodesId := make([]uint64, 0)
				for _, nodeInfo := range w.GetCurrentPeerInfo() {
					if nodeInfo.State == infos.NodeState_ONLINE {
						if nodeId != nodeInfo.RaftId {
							onlineNodesId = append(onlineNodesId, nodeInfo.RaftId)
						}
					}
				}

				m := w.GetMoon()
				reply, err := m.ListInfo(w.ctx, &moon2.ListInfoRequest{
					InfoType: infos.InfoType_TASK_INFO,
				})
				if err != nil {
					logger.Errorf("list task info err: %v", err)
					continue
				}
				cur := 0
				for _, info := range reply.BaseInfos {
					taskInfo := info.GetTaskInfo()
					if taskInfo.ScheduleSatelliteId == w.selfNodeInfo.RaftId {
						if taskInfo.Phase != infos.TaskInfo_Finished {
							// Process Migrate Task
							taskInfo.ScheduleSatelliteId = onlineNodesId[cur]
							cur = (cur + 1) % len(onlineNodesId)
							_, err = m.ProposeInfo(w.ctx, &moon2.ProposeInfoRequest{
								Operate: moon2.ProposeInfoRequest_UPDATE,
								Id:      taskInfo.TaskUuid,
								BaseInfo: &infos.BaseInfo{
									Info: &infos.BaseInfo_TaskInfo{
										TaskInfo: taskInfo,
									},
								},
							})
							if err != nil {
								logger.Errorf("migrate task info err: %v", err)
								continue
							}
							logger.Infof("migrate task info success, taskUUID: %v from node: %v to node: %v",
								taskInfo.TaskUuid, nodeId, taskInfo.ScheduleSatelliteId)
						}

					}
				}
			}
		}
	}
}

func (w *Watcher) processTask() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			m := w.GetMoon()
			reply, err := m.ListInfo(w.ctx, &moon2.ListInfoRequest{
				InfoType: infos.InfoType_TASK_INFO,
			})
			if err != nil {
				logger.Errorf("list task info err: %v", err)
				continue
			}
			for _, info := range reply.BaseInfos {
				taskInfo := info.GetTaskInfo()
				if taskInfo.ScheduleSatelliteId == w.selfNodeInfo.RaftId {
					// TODO: (refactor) process task use k8s client

					// update task info
					taskInfo.Phase = infos.TaskInfo_Finished
					_, err := m.ProposeInfo(w.ctx, &moon2.ProposeInfoRequest{
						Operate: moon2.ProposeInfoRequest_UPDATE,
						Id:      taskInfo.TaskUuid,
						BaseInfo: &infos.BaseInfo{
							Info: &infos.BaseInfo_TaskInfo{
								TaskInfo: taskInfo,
							},
						},
					})

					if err != nil {
						logger.Errorf("delete task info err: %v", err)
						continue
					}
				}
			}
		}
	}
}

func (w *Watcher) initCluster() {
	//_, err := w.moon.ProposeInfo(w.ctx, &moon2.ProposeInfoRequest{
	//	Operate: moon2.ProposeInfoRequest_ADD,
	//})
	//if err != nil {
	//	logger.Errorf("init cluster fail: %v", err)
	//}
	return
}

func (w *Watcher) proposeClusterInfo(clusterInfo *infos.ClusterInfo) {
	if !w.moon.IsLeader() {
		return
	}
	request := &moon2.ProposeInfoRequest{
		Head: &common.Head{
			Timestamp: timestamp.Now(),
			Term:      w.GetCurrentTerm(),
		},
		Operate: moon2.ProposeInfoRequest_ADD,
		Id:      strconv.FormatUint(clusterInfo.Term, 10),
		BaseInfo: &infos.BaseInfo{Info: &infos.BaseInfo_ClusterInfo{
			ClusterInfo: clusterInfo,
		}},
	}
	logger.Infof("[NEW TERM] leader: %v propose new cluster info, term: %v, node num: %v",
		w.selfNodeInfo.RaftId, request.BaseInfo.GetClusterInfo().Term,
		len(request.BaseInfo.GetClusterInfo().NodesInfo))
	_, err := w.moon.ProposeInfo(w.ctx, request)
	if err != nil {
		// TODO
		logger.Errorf("propose cluster info err: %v", err)
		return
	}
	logger.Infof("[NEW TERM] leader propose new cluster info success")
}

func (w *Watcher) nodeInfoChanged(_ infos.Information) {
	w.timerMutex.Lock()
	defer w.timerMutex.Unlock()
	if !w.moon.IsLeader() {
		return
	}
	if w.timer != nil && w.timer.Stop() {
		w.timer.Reset(w.Config.NodeInfoCommitInterval)
		return
	}
	w.timer = time.AfterFunc(w.Config.NodeInfoCommitInterval, func() {
		select {
		case <-w.ctx.Done():
			return
		default:
			clusterInfo := w.genNewClusterInfo()
			if clusterInfo == nil {
				return
			}
			w.proposeClusterInfo(clusterInfo)
		}
	})
}

func (w *Watcher) clusterInfoChanged(info infos.Information) {
	if w.currentClusterInfo.Term == uint64(0) && w.moon.IsLeader() { // first time
		go w.initCluster()
	}
	w.clusterInfoMutex.Lock()
	defer w.clusterInfoMutex.Unlock()
	logger.Infof("[NEW TERM] node %v get new term: %v, node num: %v",
		w.selfNodeInfo.RaftId, info.BaseInfo().GetClusterInfo().Term,
		len(info.BaseInfo().GetClusterInfo().NodesInfo))
	w.currentClusterInfo = *info.BaseInfo().GetClusterInfo()
}

func (w *Watcher) WaitClusterOK() (ok bool) {
	for {
		select {
		case <-w.ctx.Done():
			return false
		default:
			if w.GetCurrentClusterInfo().Term != uint64(0) && w.selfNodeInfo.RaftId != 0 {
				return true
			}
			time.Sleep(time.Second)
		}
	}
}

func (w *Watcher) SubmitGeoUnSensitiveTask(_ context.Context, request *GeoUnSensitiveTaskRequest) (*common.Result, error) {
	if w.moon.IsLeader() {
		// TODO upload file
		// propose task info to TCS
		m := w.GetMoon()
		// search one node which has the least task
		reply, err := m.ListInfo(w.ctx, &moon2.ListInfoRequest{
			InfoType: infos.InfoType_TASK_INFO,
		})
		if err != nil {
			logger.Errorf("list task info err: %v", err)
			return nil, err
		}
		nodeTaskCountMap := make(map[uint64]int)
		for _, info := range reply.BaseInfos {
			taskInfo := info.GetTaskInfo()
			nodeTaskCountMap[taskInfo.ScheduleSatelliteId]++
		}
		minTaskCount := 0
		minTaskNode := uint64(0)
		for node, count := range nodeTaskCountMap {
			if minTaskCount == 0 || count < minTaskCount {
				minTaskCount = count
				minTaskNode = node
			}
		}
		_, err = m.ProposeInfo(w.ctx, &moon2.ProposeInfoRequest{
			Operate: moon2.ProposeInfoRequest_ADD,
			Id:      request.TaskUuid,
			BaseInfo: &infos.BaseInfo{
				Info: &infos.BaseInfo_TaskInfo{
					TaskInfo: &infos.TaskInfo{
						TaskUuid:            request.TaskUuid,
						Phase:               infos.TaskInfo_Created,
						ImageName:           request.ImageName,
						ScheduleSatelliteId: minTaskNode,
					},
				},
			},
		})
		if err != nil {
			logger.Errorf("propose task info err: %v", err)
			return nil, err
		}
	} else {
		leaderInfo := w.GetCurrentClusterInfo().LeaderInfo
		conn, err := messenger.GetRpcConnByNodeInfo(leaderInfo)
		if err != nil {
			return nil, err
		}
		client := NewWatcherClient(conn)
		return client.SubmitGeoUnSensitiveTask(w.ctx, request)
	}

	return nil, nil
}

func NewWatcher(ctx context.Context, config *Config, server *messenger.RpcServer,
	m moon.InfoController, register *infos.StorageRegister) *Watcher {
	watcherCtx, cancelFunc := context.WithCancel(ctx)

	watcher := &Watcher{
		moon:         m,
		register:     register,
		Config:       config,
		selfNodeInfo: &config.SelfNodeInfo,
		ctx:          watcherCtx,
		cancelFunc:   cancelFunc,
	}
	monitor := NewMonitor(watcherCtx, watcher, server)
	watcher.Monitor = monitor
	nodeInfoStorage := watcher.register.GetStorage(infos.InfoType_NODE_INFO)
	clusterInfoStorage := watcher.register.GetStorage(infos.InfoType_CLUSTER_INFO)
	nodeInfoStorage.SetOnUpdate("watcher-"+watcher.selfNodeInfo.Uuid, watcher.nodeInfoChanged)
	clusterInfoStorage.SetOnUpdate("watcher-"+watcher.selfNodeInfo.Uuid, watcher.clusterInfoChanged)
	NewStatusReporter(ctx, watcher)

	RegisterWatcherServer(server, watcher)
	return watcher
}
