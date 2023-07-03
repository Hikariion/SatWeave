package client

import (
	"context"
	"io"
	"os"
	"path"
	"satweave/client/config"
	"satweave/client/offload"
	"satweave/messenger"
	"satweave/utils/logger"
)

type Client struct {
	offload.UnimplementedClientServer

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

	offload.RegisterClientServer(rpcServer, c)

	return c
}

func (c *Client) ReceiveFile(stream offload.Client_ReceiveFileServer) error {
	var filename string
	var file *os.File

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&offload.ReceiveFileReply{
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

func (c *Client) UploadFile(satIpAddr string, satRpcPort uint64, fileDir, filename string) {

}
