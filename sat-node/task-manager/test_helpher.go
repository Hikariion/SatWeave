package task_manager

import (
	"context"
	"fmt"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/utils/logger"
	"strconv"
	"strings"
	"time"
)

const slotNum = 1000

func GenTestTaskManager(ctx context.Context, basePath string, sunAddr string, satelliteName string, slotNum uint64,
	host string, pathNodes []string) (*TaskManager, *messenger.RpcServer) {
	port, nodeRpc := messenger.NewRandomPortRpcServer()

	cloudAddr := strings.Split(sunAddr, ":")[0]
	cloudPort, _ := strconv.Atoi(strings.Split(sunAddr, ":")[1])

	taskManagerConfig := DefaultConfig
	taskManagerConfig.SunAddr = sunAddr
	taskManagerConfig.CloudAddr = cloudAddr
	taskManagerConfig.CloudPort = uint64(cloudPort)
	taskManagerConfig.StoragePath = basePath

	taskManager := NewTaskManager(ctx, &taskManagerConfig, satelliteName, nodeRpc, slotNum, host, port, pathNodes)

	return taskManager, nodeRpc
}

func GenTestTaskManagerCluster(ctx context.Context, basePath string, num int) ([]*TaskManager, []*messenger.RpcServer, string, *sun.Sun,
	[]string) {
	sunPort, sunRpc := messenger.NewRandomPortRpcServer()
	s := sun.NewSun(sunRpc)

	go func() {
		err := sunRpc.Run()
		if err != nil {
			logger.Errorf("Run rpcServer err: %v", err)
		}
	}()
	sunAddr := "127.0.0.1:" + strconv.FormatUint(sunPort, 10)

	time.Sleep(2 * time.Second)

	var taskManagers []*TaskManager
	var rpcServers []*messenger.RpcServer

	var pathNodes []string
	for i := 0; i < num; i++ {
		pathNodes = append(pathNodes, fmt.Sprintf("satellite%d", uint64(i+1)))
	}

	for i := 0; i < num; i++ {
		taskManager, rpc := GenTestTaskManager(ctx, basePath, sunAddr, pathNodes[i], slotNum, "127.0.0.1", pathNodes)
		taskManagers = append(taskManagers, taskManager)
		rpcServers = append(rpcServers, rpc)
	}
	return taskManagers, rpcServers, sunAddr, s, pathNodes
}

func RunAllTestTaskManager(taskManagers []*TaskManager) {
	for _, t := range taskManagers {
		t.Run()
	}
}
