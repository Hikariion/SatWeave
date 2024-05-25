package sun

import "time"

type NodeStatus struct {
	NodeId         uint64
	LastReportTime int64
	Online         bool
}

type NodeStatusCollector struct {
	Nodes map[uint64]*NodeStatus
}

func (n *NodeStatusCollector) AddNodeStatus(nodeId uint64) {
	n.Nodes[nodeId] = &NodeStatus{
		NodeId: nodeId,
		Online: true,
	}
}

func (n *NodeStatusCollector) NodeReportHealth(nodeId uint64) {
	if _, ok := n.Nodes[nodeId]; ok {
		n.Nodes[nodeId].LastReportTime = time.Now().Unix()
	} else {
		n.AddNodeStatus(nodeId)
		n.Nodes[nodeId].LastReportTime = time.Now().Unix()
	}
}

func (n *NodeStatusCollector) UpdateNodeStatus() {
	now := time.Now().Unix()
	for _, v := range n.Nodes {
		if now-v.LastReportTime > 3 {
			v.Online = false
		}
	}
}

func (n *NodeStatusCollector) Run() {
	for {
		n.UpdateNodeStatus()
		time.Sleep(time.Second)
	}
}

func NewNodeStatusCollector() *NodeStatusCollector {
	return &NodeStatusCollector{
		Nodes: make(map[uint64]*NodeStatus),
	}
}
