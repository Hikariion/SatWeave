package task_manager

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"satweave/cloud/sun"
	"satweave/messenger"
	"testing"
	"time"
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
	taskManagers, rpcServers, _, s := GenTestTaskManagerCluster(ctx, basePath, nodeNum)

	for i := 0; i < nodeNum; i++ {
		go func(rpc *messenger.RpcServer) {
			err := rpc.Run()
			if err != nil {
				t.Errorf("rpc server run error: %v", err)
			}
		}(rpcServers[i])
	}

	RunAllTestTaskManager(taskManagers)

	s.PrintTaskManagerTable()

	t.Run("test sun schedule ", func(t *testing.T) {
		yamlBytes, err := os.ReadFile("./test-files/FFT_config.yaml")
		assert.NoError(t, err)

		request := &sun.SubmitJobRequest{
			JobId:         "1071dfe3-d54a-4a0e-9bcb-64d6e984781d",
			YamlByte:      yamlBytes,
			SatelliteName: "satellite1",
		}

		response, err := s.SubmitJob(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, response.Success)

		time.Sleep(time.Second * 100)
	})

}
