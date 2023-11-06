package task_manager

import (
	"context"
	"github.com/stretchr/testify/assert"
	"satweave/messenger"
	"satweave/utils/common"
	"satweave/utils/logger"
	"testing"
)

func TestTaskManager(t *testing.T) {
	t.Run("task manager test", func(t *testing.T) {
		testTaskManager(t)
	})
}

func testTaskManager(t *testing.T) {
	basePath := "./sat-data/task-manager-test"
	nodeNum := 10
	ctx := context.Background()
	var taskManagers []*TaskManager
	var rpcServers []*messenger.RpcServer

	// Run Sun
	taskManagers, rpcServers, _, sun := GenTestTaskManagerCluster(ctx, basePath, nodeNum)

	for i := 0; i < nodeNum; i++ {
		go func(rpc *messenger.RpcServer) {
			err := rpc.Run()
			if err != nil {
				t.Errorf("rpc server run error: %v", err)
			}
		}(rpcServers[i])
	}

	RunAllTestTaskManager(taskManagers)

	sun.PrintTaskManagerTable()

	t.Run("test sun schedule ", func(t *testing.T) {
		userTasks, err := common.ReadUserDefinedTasks("./test-files/task_config.yaml")
		assert.NoError(t, err)
		assert.NotEmpty(t, userTasks)
		//logger.Infof("%v", userTasks)
		logicalTask, err := common.ConvertUserTaskWrapperToLogicTasks(userTasks)
		assert.NoError(t, err)
		assert.NotEmpty(t, logicalTask)

		logicalTaskMap, executeTaskMap, err := sun.Scheduler.Schedule(0, logicalTask)
		assert.NoError(t, err)

		logger.Infof("logicalTaskMap: %v", logicalTaskMap)
		logger.Infof("executeTaskMap: %v", executeTaskMap)
	})
}
