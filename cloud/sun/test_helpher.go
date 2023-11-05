package sun

import (
	"context"
	"satweave/messenger"
	"satweave/cloud/sun"
	"satweave/sat-node/task-manager"
	"satweave/utils/logger"
	"strconv"
	"time"
)

func GenTestTaskManager() (*task_manager.TaskManager, *messenger.RpcServer) {
	return nil, nil
}

func GenTestTaskManagerCluster(ctx context.Context, basePath string, num int) ([]*task_manager.TaskManager, []*messenger.RpcServer, string)  {
	sunPort, sunRpc := messenger.NewRandomPortRpcServer()
	sun.NewSun(sunRpc)

	go func() {
		err := sunRpc.Run()
		if err != nil {
			logger.Errorf("Run rpcServer err: %v", err)
		}
	}()
	sunAddr := "127.0.0.1:" + strconv.FormatUint(sunPort, 10)

	time.Sleep(2 * time.Second)

	var taskManagers []*task_manager.TaskManager
	var rpcServers []*messenger.RpcServer
	for

}