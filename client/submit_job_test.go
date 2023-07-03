package client

import (
	"context"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"satweave/client/config"
	"satweave/messenger"
	"satweave/sat-node/worker"
	"satweave/shared/service"
	"satweave/utils/common"
	"satweave/utils/logger"
	"strconv"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	t.Run("upload file", func(t *testing.T) {
		testUploadFile(t)
	})
}

// 测试客户端上传文件功能
func testUploadFile(t *testing.T) {
	// 创建worker的rpc
	ctx := context.Background()
	satPort, satRpcServer := messenger.NewRandomPortRpcServer()
	workerConfig := &worker.Config{
		AttachmentStoragePath: "/Users/zhangjh/code/SatWeave/client/satweave-data/attachment/",
		OutputPath:            "/Users/zhangjh/code/SatWeave/client/satweave-data/output/",
	}
	// 创建StoragePath
	err := common.InitPath(workerConfig.AttachmentStoragePath)
	if err != nil {
		t.Errorf("InitPath err: %v", err)
	}
	err = common.InitPath(workerConfig.OutputPath)
	if err != nil {
		t.Errorf("InitPath err: %v", err)
	}
	w := worker.NewWorker(ctx, satRpcServer, workerConfig)
	go satRpcServer.Run()

	// client 向 worker 上传文件
	clientPort, clientRpcServer := messenger.NewRandomPortRpcServer()
	clientConfig := &config.ClientConfig{
		ClientIp:                     "127.0.0.1",
		RpcPort:                      clientPort,
		FileFromSatelliteStoragePath: "/Users/zhangjh/code/SatWeave/client/satweave-data/back-files/",
	}
	err = common.InitPath(clientConfig.FileFromSatelliteStoragePath)
	if err != nil {
		t.Errorf("InitPath err: %v", err)
	}

	client := NewClient(ctx, clientConfig, clientRpcServer)
	go clientRpcServer.Run()

	err = client.UploadFile("127.0.0.1", satPort, "./satweave-data/source/", "bus.jpg")
	assert.NoError(t, err)
	err = client.UploadFile("127.0.0.1", satPort, "./satweave-data/source/", "zidane.jpg")
	assert.NoError(t, err)

	md5SumOri, err := calcFileHash("./satweave-data/source/bus.jpg")
	md5SumUpload, err := calcFileHash("./satweave-data/attachment/bus.jpg")
	assert.Equal(t, md5SumOri, md5SumUpload)

	md5SumOri, err = calcFileHash("./satweave-data/source/zidane.jpg")
	md5SumUpload, err = calcFileHash("./satweave-data/attachment/zidane.jpg")
	assert.Equal(t, md5SumOri, md5SumUpload)

	// 测试提交任务

	job1 := &service.Job{
		ClientIp:   client.config.ClientIp,
		ClientPort: client.config.RpcPort,
		// 时间戳作为 JobId
		JobId:     strconv.FormatInt(time.Now().Unix(), 10),
		ImageName: "harbor.act.buaa.edu.cn/satweave/satyolov5",
		// 任务附件的文件名
		Attachment: "bus.jpg",
		ResultName: "bus.txt",
		Command:    "python3 detect.py --source ./data/images/bus.jpg --save-txt --nosave",
	}

	job2 := &service.Job{
		ClientIp:   client.config.ClientIp,
		ClientPort: client.config.RpcPort,
		// 时间戳作为 JobId
		JobId: strconv.FormatInt(time.Now().Unix(), 10),
		// 任务附件的文件名
		Attachment: "zidane.jpg",
		ResultName: "bus.txt",
		ImageName:  "harbor.act.buaa.edu.cn/satweave/satyolov5",
		Command:    "python3 detect.py --source ./data/images/zidane.jpg --save-txt --nosave",
	}

	err = client.submitJob(ctx, "127.0.0.1", satPort, job1)
	assert.NoError(t, err)
	err = client.submitJob(ctx, "127.0.0.1", satPort, job2)

	readJob1 := <-w.JobQueue
	readJob2 := <-w.JobQueue

	assert.Equal(t, job1, readJob1)
	assert.Equal(t, job2, readJob2)

	//w.Run()
	err = w.ExecuteJob(ctx, job1)
	assert.NoError(t, err)
	err = w.ExecuteJob(ctx, job2)
	assert.NoError(t, err)

	time.Sleep(5 * time.Second)

	clientRpcServer.Stop()
	satRpcServer.Stop()

	// 删除文件
	//err = os.Remove("./satweave-data/attachment/bus.jpg")
	//err = os.Remove("./satweave-data/attachment/zidane.jpg")
}

// 测试提交任务的功能
func testSubmitJob(t *testing.T) {

}

// 计算文件的哈希值
func calcFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatalf("Open file err: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		logger.Fatalf("Copy file err: %v", err)
	}

	hashInBytes := hash.Sum(nil)
	hashString := string(hashInBytes)
	return hashString, nil
}
