package watcher

import (
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/infos"
	"satweave/utils/logger"
	"sort"
	"sync"
	"time"
)

type Monitor interface {
	MonitorServer
	Run()
	GetAllNodeReports() []*NodeStatusReport
	GetNodeReport(nodeId uint64) *NodeStatusReport
	GetEventChannel() <-chan *Event
	Register(name string, reporter Reporter) error
	Stop()
}

type Event struct {
	Report *NodeStatusReport
}

type ReportType int32

const (
	ReportTypeADD ReportType = iota
	ReportTypeUPDATE
	ReportTypeDELETE
)

type Report struct {
	ReportType ReportType
	NodeReport *NodeStatusReport
}

type Reporter interface {
	IsChanged() bool
	GetReports() []Report
}

type NodeMonitor struct {
	UnimplementedMonitorServer
	ctx           context.Context
	cancel        context.CancelFunc
	timer         *time.Ticker
	clusterReport *ClusterReport

	nodeStatusMap sync.Map
	reportTimers  sync.Map

	selfNodeStatusMutex sync.Mutex
	selfNodeStatus      *NodeStatusReport
	watcher             *Watcher

	reportersMap sync.Map
	eventChannel chan *Event
}

func (m *NodeMonitor) Register(name string, reporter Reporter) error {
	if _, ok := m.reportersMap.Load(name); ok {
		return errors.New("reporter already registered")
	}
	m.reportersMap.Store(name, reporter)
	return nil
}

// CollectNodes 收集 Cluster 中所有 Node 的节点信息
func (m *NodeMonitor) collectNodes() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.timer.C:
			// 检查自己是否为 leader
			//logger.Debugf("Node: %v check if need collect nodes", m.watcher.GetSelfInfo().GetID())
			if !m.watcher.moon.IsLeader() {
				//logger.Debugf("Node: %v not leader", m.watcher.GetSelfInfo().GetID())
				time.Sleep(time.Second * 3)
				continue
			}
			logger.Debugf("Node: %v start collect node report", m.watcher.GetSelfInfo().GetID())

			m.watcher.GetCurrentClusterInfo()
			nodeInfoStorage := m.watcher.register.GetStorage(infos.InfoType_NODE_INFO)
			nodeInfos, err := nodeInfoStorage.GetAll()
			if err != nil {
				logger.Errorf("get nodeInfo from nodeInfoStorage fail: %v", err)
				//return
			}
			logger.Tracef("Node: %v start send collect request, node size: %v", m.watcher.GetSelfInfo().GetID(), len(nodeInfos))

			// 拉取其他节点信息
			for _, nodeInfo := range nodeInfos {
				go func(inform infos.Information) {
					nodeInfo := inform.BaseInfo().GetNodeInfo()
					conn, _ := messenger.GetRpcConnByNodeInfo(nodeInfo)
					client := NewMonitorClient(conn)
					ctxWithTimeOut, cancel := context.WithTimeout(m.ctx, time.Second*10)
					defer cancel()
					report, err := client.Get(ctxWithTimeOut, &emptypb.Empty{})
					if err != nil {
						// 获取节点信息超时，汇报错误信息
						logger.Warningf("get node status timeout, nodeId: %v, err: %v", nodeInfo.RaftId, err.Error())
						v, ok := m.nodeStatusMap.Load(nodeInfo.RaftId)
						if !ok {
							return
						}
						r := v.(*NodeStatusReport)
						if r.State == infos.NodeState_OFFLINE {
							return
						}
						r.State = infos.NodeState_OFFLINE
						m.nodeStatusMap.Store(nodeInfo.RaftId, r)
						m.eventChannel <- &Event{
							Report: r,
						}
						return
					}

					if _, ok := m.nodeStatusMap.Load(report.NodeId); !ok {
						// first time online
						m.eventChannel <- &Event{
							Report: report,
						}
					}
					m.nodeStatusMap.Store(report.NodeId, report)

					logger.Tracef("Node: %v get node report from: %v", m.watcher.GetSelfInfo().GetID(), nodeInfo.RaftId)
				}(nodeInfo)
			}

		}
	}
}

// Report is a rpc func to get the node status report.
// it called by all node (include leader self).
func (m *NodeMonitor) Report(_ context.Context, report *NodeStatusReport) (*common.Result, error) {
	if !m.watcher.GetMoon().IsLeader() {
		// only leader can be runReport
		logger.Infof("node %v is not leader, can't runReport", m.watcher.GetSelfInfo().GetID())
		return nil, errors.New("not leader")
	}
	if val, ok := m.reportTimers.Load(report.NodeId); ok {
		logger.Tracef("reset timer for node %v", report.NodeId)
		t := val.(*time.Timer)
		if !t.Stop() {
			<-t.C
		}
		t.Reset(time.Second * 3)
	} else {
		logger.Debugf("create timer for node %v", report.NodeId)
		m.reportTimers.Store(report.NodeId, time.AfterFunc(time.Second*3, func() {
			select {
			case <-m.ctx.Done():
				return
			default:
			}
			logger.Warningf("get node status timeout, nodeId: %v", report.NodeId)
			v, _ := m.nodeStatusMap.Load(report.NodeId)
			r := v.(*NodeStatusReport)
			r.State = infos.NodeState_OFFLINE
			m.nodeStatusMap.Store(report.NodeId, r)
			m.eventChannel <- &Event{
				Report: r,
			}
		}))
	}
	if _, ok := m.nodeStatusMap.Load(report.NodeId); !ok {
		// first time online
		m.eventChannel <- &Event{
			Report: report,
		}
	}
	m.nodeStatusMap.Store(report.NodeId, report)
	return &common.Result{}, nil
}

// GetEventChannel returns the event channel.
// Event channel will send event when the node status changed. (like node online, offline, etc.)
func (m *NodeMonitor) GetEventChannel() <-chan *Event {
	return m.eventChannel
}

// GetAllNodeReports returns all node status reports.
func (m *NodeMonitor) GetAllNodeReports() []*NodeStatusReport {
	var nodeStatusList []*NodeStatusReport
	m.nodeStatusMap.Range(func(key, value interface{}) bool {
		nodeStatusList = append(nodeStatusList, value.(*NodeStatusReport))
		return true
	})
	sort.Slice(nodeStatusList, func(i, j int) bool {
		return nodeStatusList[i].NodeId < nodeStatusList[j].NodeId
	})
	return nodeStatusList
}

func (m *NodeMonitor) GetNodeReport(nodeID uint64) *NodeStatusReport {
	if val, ok := m.nodeStatusMap.Load(nodeID); ok {
		return val.(*NodeStatusReport)
	}
	return nil
}

func (m *NodeMonitor) GetClusterReport(context.Context, *emptypb.Empty) (*ClusterReport, error) {
	reports := m.GetAllNodeReports()
	clusterState := ClusterReport_HEALTH_OK

	// 获取最新集群信息
	clusterInfo := m.watcher.GetCurrentClusterInfo()

	return &ClusterReport{
		State:       clusterState,
		ClusterInfo: &clusterInfo,
		Nodes:       reports,
	}, nil
}

// Get return NodeStatusReport of node self
// Get 返回自身的 Status 信息，用于拉取模型
func (m *NodeMonitor) Get(context.Context, *emptypb.Empty) (*NodeStatusReport, error) {
	m.selfNodeStatusMutex.Lock()
	defer m.selfNodeStatusMutex.Unlock()
	if m.selfNodeStatus == nil {
		m.selfNodeStatusMutex.Unlock()
		m.collectReports()
		m.selfNodeStatusMutex.Lock()
	}
	return m.selfNodeStatus, nil
}

// collectReports 收集自身的 status 信息
func (m *NodeMonitor) collectReports() {
	m.selfNodeStatusMutex.Lock()
	defer m.selfNodeStatusMutex.Unlock()
	m.reportersMap.Range(func(key, value interface{}) bool {
		reporter := value.(Reporter)
		if reporter.IsChanged() == false {
			return true
		}
		reports := reporter.GetReports()

		for _, report := range reports {
			if report.NodeReport != nil {
				m.selfNodeStatus = report.NodeReport
			}
		}
		return true
	})
}

func (m *NodeMonitor) runReport() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.timer.C:
			m.collectReports()
			leaderID := m.watcher.GetMoon().GetLeaderID()
			if leaderID == 0 {
				continue
			}
		}
	}
}

func (m *NodeMonitor) Run() {
	m.timer = time.NewTicker(time.Second * 1)
	go m.collectNodes()
	go m.runReport()
}

func (m *NodeMonitor) Stop() {
	m.cancel()
	m.timer.Stop()
}

func NewMonitor(ctx context.Context, w *Watcher, rpcServer *messenger.RpcServer) Monitor {
	ctx, cancel := context.WithCancel(ctx)
	monitor := &NodeMonitor{
		ctx:           ctx,
		cancel:        cancel,
		nodeStatusMap: sync.Map{},
		watcher:       w,
		eventChannel:  make(chan *Event, 5),
	}
	RegisterMonitorServer(rpcServer, monitor)
	return monitor
}
