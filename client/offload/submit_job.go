package offload

import (
	"context"
	"google.golang.org/grpc"
	"io"
	"os"
	"path"
	"satweave/sat-node/worker"
	"satweave/shared/service"
	"satweave/utils/logger"
)

// 客户端上传一个任务附件给卫星
func uploadFile(satIp, satRpcPort, fileDir, filename string) {
	// 卫星的rpc地址， 卫星ip + 端口
	conn, err := grpc.Dial(satIp+":"+satRpcPort, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := worker.NewWorkerClient(conn)

	file, err := os.Open(path.Join(fileDir, filename))
	if err != nil {
		logger.Fatalf("Error while opening file: %v", err)
	}
	defer file.Close()

	stream, err := c.UploadAttachment(context.Background())
	if err != nil {
		logger.Fatalf("Error while opening stream: %v", err)
	}

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Fatalf("Error while reading chunk: %v", err)
		}
		stream.Send(&service.Chunk{
			Filename: filename,
			Data:     buffer[:n],
		})
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		logger.Fatalf("Error while receiving response: %v", err)
	}

	logger.Infof("Response: %v", res)
}
