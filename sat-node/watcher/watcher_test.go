package watcher

import (
	"context"
	"os"
	"satweave/messenger"
	"satweave/sat-node/infos"
	moon2 "satweave/shared/moon"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	t.Run("watcher with real moon", func(t *testing.T) {
		testWatcher(t)
	})
}

func testWatcher(t *testing.T) {
	basePath := "./sat-data/watcher-test"
	nodeNum := 9
	ctx := context.Background()
	var watchers []*Watcher
	var rpcServers []*messenger.RpcServer
	var leader uint64

	// Run Sun
	watchers, rpcServers, _ = GenTestWatcherCluster(ctx, basePath, nodeNum)

	for i := 0; i < nodeNum; i++ {
		go func(rpc *messenger.RpcServer) {
			err := rpc.Run()
			if err != nil {
				t.Errorf("rpc server run error: %v", err)
			}
		}(rpcServers[i])
	}
	// Run First half
	firstRunNum := nodeNum/2 + 1

	RunAllTestWatcher(watchers[:firstRunNum])
	WaitAllTestWatcherOK(watchers[:firstRunNum])

	moon := watchers[0].moon
	_, err := moon.ProposeInfo(ctx, &moon2.ProposeInfoRequest{
		Operate: moon2.ProposeInfoRequest_ADD,
		Id:      "test_task",
		BaseInfo: &infos.BaseInfo{
			Info: &infos.BaseInfo_TaskInfo{
				TaskInfo: &infos.TaskInfo{
					UserIp: "127.0.0.1",
				},
			},
		},
	})
	if err != nil {
		t.Errorf("propose info error: %v", err)
	}

	t.Run("add left half", func(t *testing.T) {
		RunAllTestWatcher(watchers[firstRunNum : len(watchers)-1])
		WaitAllTestWatcherOK(watchers[:len(watchers)-1])
		// test node
	})

	t.Cleanup(func() {
		for i := 0; i < len(watchers); i++ {
			watchers[i].moon.Stop()
			rpcServers[i].Stop()
		}
		_ = os.RemoveAll(basePath)
	})

	t.Logf("watcher leader info %v", watchers[0].moon.GetLeaderID())

	t.Run("remove a node", func(t *testing.T) {
		watchers[len(watchers)-1].moon.Stop()
		watchers[len(watchers)-1].cancelFunc()
		rpcServers[len(watchers)-1].Stop()
		watchers = watchers[:len(watchers)-1]
		rpcServers = rpcServers[:len(rpcServers)-1]
		WaitAllTestWatcherOK(watchers[:len(watchers)-1])
	})

	t.Run("leader fail", func(t *testing.T) {
		leader = watchers[0].moon.GetLeaderID()
		t.Logf("old leader %v", leader)
		watchers[leader-1].moon.Stop()
		watchers[leader-1].cancelFunc()
		rpcServers[leader-1].Stop()
		time.Sleep(5 * time.Second)
		watchers = append(watchers[:leader-1], watchers[leader:len(watchers)-1]...)
		rpcServers = append(rpcServers[:leader-1], rpcServers[leader:len(watchers)-1]...)
		WaitAllTestWatcherOK(watchers[:len(watchers)-1])
		leader = watchers[0].moon.GetLeaderID()
		t.Logf("new leader %v", leader)
	})
}
