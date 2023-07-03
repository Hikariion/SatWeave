package client

import (
	"context"
	"io"
	"os"
	"path"
	"satweave/client/config"
	"satweave/messenger"
	"satweave/shared/client"
	"satweave/shared/service"
	"satweave/shared/worker"
	"satweave/utils/logger"
)

type Client struct {
	client.UnimplementedClientServer

	ctx    context.Context
	cancel context.CancelFunc

	config *config.ClientConfig
}

func NewClient(ctx context.Context, config *config.ClientConfig, rpcServer *messenger.RpcServer) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		ctx:    ctx,
		cancel: cancel,
		config: config,
	}

	client.RegisterClientServer(rpcServer, c)

	return c
}

func (c *Client) ReceiveFile(stream client.Client_ReceiveFileServer) error {
	var filename string
	var file *os.File

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&client.ReceiveFileReply{
				Success: true,
			})
		}
		if err != nil {
			logger.Errorf("Error while reading chunks: %v", err)
			return err
		}

		if filename == "" {
			filename = chunk.GetFilename()
			// 创建附件的存储路径
			file, err = os.Create(path.Join(c.config.FileFromSatelliteStoragePath, filename))
			if err != nil {
				return err
			}
			defer file.Close()
		}

		_, writeErr := file.Write(chunk.GetData())
		if writeErr != nil {
			logger.Errorf("Error while writing to file: %v", writeErr)
			return writeErr
		}
	}
}

// UploadFile 客户端上传一个任务附件给卫星
func (c *Client) UploadFile(satIpAddr string, satRpcPort uint64, fileDir, filename string) error {
	// 卫星的rpc地址， 卫星ip + 端口
	conn, err := messenger.GetRpcConn(satIpAddr, satRpcPort)
	if err != nil {
		logger.Errorf("did not connect: %v", err)
		return err
	}
	defer conn.Close()
	cc := worker.NewWorkerClient(conn)

	file, err := os.Open(path.Join(fileDir, filename))
	if err != nil {
		logger.Errorf("Error while opening file: %v", err)
		return err
	}
	defer file.Close()

	stream, err := cc.UploadAttachment(context.Background())
	if err != nil {
		logger.Errorf("Error while opening stream: %v", err)
		return err
	}

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Errorf("Error while reading chunk: %v", err)
			return err
		}
		stream.Send(&worker.Chunk{
			Filename: filename,
			Data:     buffer[:n],
		})
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		logger.Errorf("Error while receiving response: %v", err)
		return err
	}

	logger.Infof("Response: %v", res)

	return nil
}

// 客户端提交一个Job给卫星
func (c *Client) submitJob(ctx context.Context, satIpAddr string, satRpcPort uint64, job *service.Job) error {
	// 卫星的rpc地址， 卫星ip + 端口
	conn, err := messenger.GetRpcConn(satIpAddr, satRpcPort)
	if err != nil {
		logger.Fatalf("did not connect: %v", err)
		return err
	}
	defer conn.Close()
	cc := worker.NewWorkerClient(conn)

	reply, err := cc.SubmitJob(ctx, &worker.SubmitJobRequest{
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
