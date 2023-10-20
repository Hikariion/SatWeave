package watcher

import (
	"context"
	"os"
	"satweave/messenger"
	"satweave/sat-node/infos"
	moon2 "satweave/shared/moon"
	"testing"
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
		Operate:  moon2.ProposeInfoRequest_ADD,
		Id:       "test_task",
		BaseInfo: &infos.BaseInfo{},
	})
	if err != nil {
		t.Errorf("propose bucket error: %v", err)
	}

	t.Run("add left half", func(t *testing.T) {
		RunAllTestWatcher(watchers[firstRunNum:])
		WaitAllTestWatcherOK(watchers)
		// test node
	})

	t.Cleanup(func() {
		for i := 0; i < nodeNum; i++ {
			watchers[i].moon.Stop()
			rpcServers[i].Stop()
		}
		_ = os.RemoveAll(basePath)
	})

	t.Run("remove a node", func(t *testing.T) {
		watchers[nodeNum-1].moon.Stop()
		watchers[nodeNum-1].cancelFunc()
		rpcServers[nodeNum-1].Stop()
		WaitAllTestWatcherOK(watchers[:nodeNum-1])
	})
}
