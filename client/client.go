package client

import (
	"context"
	"satweave/client/config"
	agent "satweave/client/info-agent"
	"satweave/messenger"
	"satweave/sat-node/infos"
	"satweave/shared/moon"
	"satweave/utils/errno"
	"satweave/utils/logger"
	"sync"
)

type Client struct {
	ctx    context.Context
	cancel context.CancelFunc

	config    *config.ClientConfig
	InfoAgent *agent.InfoAgent

	mutex sync.RWMutex
}

func New(config *config.ClientConfig) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	clusterInfo := &infos.ClusterInfo{
		NodesInfo: []*infos.NodeInfo{
			{
				RaftId:  0,
				IpAddr:  config.NodeAddr,
				RpcPort: config.NodePort,
				State:   infos.NodeState_ONLINE,
			},
		},
	}
	infoAgent, err := agent.NewInfoAgent(ctx, clusterInfo, config.CloudAddr, config.CloudPort, config.ConnectType)
	if err != nil {
		cancel()
		return nil, err
	}
	return &Client{
		ctx:       ctx,
		config:    config,
		cancel:    cancel,
		InfoAgent: infoAgent,
	}, nil
}

func (client *Client) GetMoon() (moon.MoonClient, uint64, error) {
	for _, nodeInfo := range client.InfoAgent.GetCurClusterInfo().NodesInfo {
		if nodeInfo.State != infos.NodeState_OFFLINE {
			continue
		}
		conn, err := messenger.GetRpcConnByNodeInfo(nodeInfo)
		if err != nil {
			logger.Errorf("get rpc connection to node: %v fail: %v", nodeInfo.RaftId, err.Error())
			continue
		}
		moonClient := moon.NewMoonClient(conn)
		return moonClient, nodeInfo.RaftId, nil
	}
	return nil, 0, errno.ConnectionIssue
}

func (client *Client) GetClusterOperator() Operator {
	return NewClusterOperator(client.ctx, client)
}
