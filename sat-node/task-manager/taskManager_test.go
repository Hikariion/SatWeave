package task_manager

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/utils/logger"
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

	taskManagers, rpcServers, _, s, pathNodes := GenTestTaskManagerCluster(ctx, basePath, nodeNum)

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

		//JobId := generator.GetJobIdGeneratorInstance().Next()

		request := &sun.SubmitJobRequest{
			JobId:         "cb5f6a1d-16bc-42c1-882a-e62bfd56ea3c",
			YamlByte:      yamlBytes,
			SatelliteName: "satellite1",
			PathNodes:     pathNodes,
		}

		response, err := s.SubmitJob(context.Background(), request)

		assert.NoError(t, err)
		assert.True(t, response.Success)

		time.Sleep(time.Second * 100)
	})
}

func TestReal(t *testing.T) {
	yamlBytes, err := os.ReadFile("./test-files/FFT_config.yaml")

	pathNodes := make([]string, 0)

	pathNodes = append(pathNodes, "satellite-0", "satellite-1", "satellite-2")

	request := &sun.SubmitJobRequest{
		JobId:         "cb5f6a1d-16bc-42c1-882a-e62bfd56ea3c",
		YamlByte:      yamlBytes,
		SatelliteName: "satellite1",
		PathNodes:     pathNodes,
	}

	jsonData, err := json.Marshal(request)
	assert.NoError(t, err)

	logger.Infof("json data: %v", string(jsonData))
}
