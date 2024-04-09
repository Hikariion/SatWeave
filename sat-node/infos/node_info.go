package infos

import (
	"github.com/google/uuid"
	"net"
	"strconv"
	"strings"
)

func (m *NodeInfo) GetInfoType() InfoType {
	return InfoType_NODE_INFO
}

func (m *NodeInfo) BaseInfo() *BaseInfo {
	return &BaseInfo{Info: &BaseInfo_NodeInfo{NodeInfo: m}}
}

func (m *NodeInfo) GetID() string {
	return strconv.FormatUint(m.RaftId, 10)
}

func getSelfIpAddr() (error, string) {
	selfIp := "127.0.0.1"
	interfaces, err := net.Interfaces()
	if err != nil {
		return err, selfIp
	}
	for _, inter := range interfaces {
		if strings.Contains(inter.Name, "docker") {
			continue
		}
		if strings.Contains(inter.Name, "lo") {
			continue
		}
		if strings.Contains(inter.Name, "br") {
			continue
		}
		address, _ := inter.Addrs()
		for _, addr := range address {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					selfIp = ipnet.IP.String()
				}
			}
		}
	}
	return nil, selfIp
}

// NewSelfInfo Generate new nodeInfo
// ** just for test **
func NewSelfInfo(raftID uint64, ipaddr string, rpcPort uint64) *NodeInfo {
	err, ip := getSelfIpAddr()
	if err != nil {
		ip = ipaddr
	}
	selfInfo := &NodeInfo{
		RaftId:   raftID,
		Uuid:     uuid.New().String(),
		IpAddr:   ip,
		RpcPort:  rpcPort,
		Capacity: 10,
	}
	return selfInfo
}
