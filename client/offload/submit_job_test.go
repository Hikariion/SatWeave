package offload

import (
	"context"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"satweave/messenger"
	"satweave/sat-node/worker"
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
	worker.NewWorker(ctx, rpcServer, workerConfig)
	go rpcServer.Run()

	// client 向 worker 上传文件

	// 把 port 转为字符串
	portStr := strconv.Itoa(int(port))

	uploadFile("127.0.0.1", portStr, "./satweave-data/source/", "bus.jpg")

	time.Sleep(3 * time.Second)

	md5SumOri, err := calcFileHash("./satweave-data/source/bus.jpg")
	md5SumUpload, err := calcFileHash("./satweave-data/attachment/bus.jpg")
	assert.Equal(t, md5SumOri, md5SumUpload)

	rpcServer.Stop()

	// 删除文件
	err = os.Remove("./satweave-data/attachment/bus.jpg")
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
