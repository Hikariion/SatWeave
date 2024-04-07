package client

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"satweave/messenger"
	"satweave/sat-node/watcher"
	"sort"
)

type Operator interface {
	State() (string, error)
	Info() (interface{}, error)
}

type ClusterOperator struct {
	client *Client
}

func NewClusterOperator(ctx context.Context, client *Client) *ClusterOperator {
	return &ClusterOperator{
		client: client,
	}
}

func (c *ClusterOperator) Info() (interface{}, error) {
	leaderInfo := c.client.InfoAgent.GetCurClusterInfo().LeaderInfo
	conn, err := messenger.GetRpcConnByNodeInfo(leaderInfo)
	if err != nil {
		return "", err
	}
	monitor := watcher.NewMonitorClient(conn)
	report, err := monitor.GetClusterReport(context.Background(), nil)
	return report, err
}

func (c *ClusterOperator) State() (string, error) {
	clusterInfo := c.client.InfoAgent.GetCurClusterInfo()
	leaderInfo := clusterInfo.LeaderInfo
	conn, err := messenger.GetRpcConnByNodeInfo(leaderInfo)
	if err != nil {
		return "", err
	}
	monitor := watcher.NewMonitorClient(conn)
	report, err := monitor.GetClusterReport(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	sort.Slice(report.Nodes, func(i, j int) bool {
		return report.Nodes[i].NodeId < report.Nodes[j].NodeId
	})
	state, _ := protoToJson(report)
	s := struct {
		Term     uint64
		LeaderID uint64
	}{
		Term:     clusterInfo.Term,
		LeaderID: clusterInfo.LeaderInfo.RaftId,
	}

	base, _ := interfaceToJson(s)

	return base + "\n" + state, nil
}

func interfaceToJson(data interface{}) (string, error) {
	// convert interface to json
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "marshal data error", err
	}
	var pretty bytes.Buffer
	err = json.Indent(&pretty, jsonData, "", "  ")
	if err != nil {
		return "indent data error", err
	}
	return pretty.String(), nil
}

func protoToJson(pb proto.Message) (string, error) {
	// convert proto to json
	marshaller := jsonpb.Marshaler{}
	jsonData, err := marshaller.MarshalToString(pb)
	if err != nil {
		return "marshal data error", err
	}
	var pretty bytes.Buffer
	err = json.Indent(&pretty, []byte(jsonData), "", "  ")
	if err != nil {
		return "indent data error", err
	}
	return pretty.String(), nil
}
