package offload

import (
	"context"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
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
	port, rpcServer := messenger.NewRandomPortRpcServer()
	workerConfig := &worker.Config{
		AttachmentStoragePath: "./satweave-data/attachment/",
	}
	// 创建StoragePath
	err := common.InitPath(workerConfig.AttachmentStoragePath)
	if err != nil {
		t.Errorf("InitPath err: %v", err)
	}
	w := worker.NewWorker(ctx, rpcServer, workerConfig)
	go rpcServer.Run()

	// client 向 worker 上传文件

	uploadFile("127.0.0.1", port, "./satweave-data/source/", "bus.jpg")
	uploadFile("127.0.0.1", port, "./satweave-data/source/", "zidane.jpg")

	time.Sleep(3 * time.Second)

	md5SumOri, err := calcFileHash("./satweave-data/source/bus.jpg")
	md5SumUpload, err := calcFileHash("./satweave-data/attachment/bus.jpg")
	assert.Equal(t, md5SumOri, md5SumUpload)

	md5SumOri, err = calcFileHash("./satweave-data/source/zidane.jpg")
	md5SumUpload, err = calcFileHash("./satweave-data/attachment/zidane.jpg")
	assert.Equal(t, md5SumOri, md5SumUpload)

	// 测试提交任务

	job1 := &service.Job{
		ClientIp: "192.168.105.134",
		// 时间戳作为 JobId
		JobId: strconv.FormatInt(time.Now().Unix(), 10),
		// 任务附件的文件名
		Attachment:     "bus.jpg",
		AttachmentSize: 100,
		Priority:       1,
	}

	job2 := &service.Job{
		ClientIp: "192.168.105.135",
		// 时间戳作为 JobId
		JobId: strconv.FormatInt(time.Now().Unix(), 10),
		// 任务附件的文件名
		Attachment:     "zidane.jpg",
		AttachmentSize: 100,
		Priority:       1,
	}

	err = submitJob(ctx, "127.0.0.1", port, job1)
	assert.NoError(t, err)
	err = submitJob(ctx, "127.0.0.1", port, job2)

	readJob1 := <-w.JobQueue
	readJob2 := <-w.JobQueue

	assert.Equal(t, job1, readJob1)
	assert.Equal(t, job2, readJob2)

	rpcServer.Stop()

	// 删除文件
	err = os.Remove("./satweave-data/attachment/bus.jpg")
	err = os.Remove("./satweave-data/attachment/zidane.jpg")
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
