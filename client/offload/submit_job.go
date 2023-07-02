package offload

import (
	"context"
	"io"
	"os"
	"path"
	"satweave/messenger"
	"satweave/sat-node/worker"
	"satweave/shared/service"
	"satweave/utils/logger"
)

// 客户端上传一个任务附件给卫星
func uploadFile(satIpAddr string, satRpcPort uint64, fileDir, filename string) {
	// 卫星的rpc地址， 卫星ip + 端口
	conn, err := messenger.GetRpcConn(satIpAddr, satRpcPort)
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

// 客户端提交一个Job给卫星
func submitJob(ctx context.Context, satIpAddr string, satRpcPort uint64, job *service.Job) error {
	// 卫星的rpc地址， 卫星ip + 端口
	conn, err := messenger.GetRpcConn(satIpAddr, satRpcPort)
	if err != nil {
		logger.Fatalf("did not connect: %v", err)
		return err
	}
	defer conn.Close()
	c := worker.NewWorkerClient(conn)

	reply, err := c.SubmitJob(ctx, &worker.SubmitJobRequest{
		Job: job,
	})
	if err != nil {
		logger.Errorf("submit job to satellite ip %v, error: %v", satIpAddr, err)
		return err
	}

	if reply.Success {
		logger.Infof("submit job to satellite ip %v, success", satIpAddr)
	} else {
		logger.Errorf("submit job to satellite ip %v, failed", satIpAddr)
	}
	return err
}
