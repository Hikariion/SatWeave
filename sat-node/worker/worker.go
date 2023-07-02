package worker

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
	"satweave/messenger"
	"satweave/shared/service"
	"satweave/utils/logger"
)

// 该模块用于执行任务

type Worker struct {
	UnimplementedWorkerServer

	ctx    context.Context
	cancel context.CancelFunc

	// 任务队列, 用于接收任务
	JobQueue chan service.Job
	// 配置
	config *Config
}

// UploadAttachment 接收客户端上传的任务附件
func (w *Worker) UploadAttachment(stream Worker_UploadAttachmentServer) error {
	var filename string
	var file *os.File

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&service.UploadAttachmentReply{
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
			file, err = os.Create(w.config.AttachmentStoragePath + filename)
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

// SubmitJob 处理客户端提交任务
func (w *Worker) SubmitJob(ctx context.Context, request *SubJobRequest) (*SubmitJobReply, error) {
	// 接收任务，将附件存储到本地
	// 将任务添加到任务队列
	return nil, status.Errorf(codes.Unimplemented, "method SubmitJob not implemented")
}

// 执行任务
func (w *Worker) executeJob(job service.Job) {
	// 从任务队列里取任务
	// 执行任务
	// 将任务结果返回给客户端
}

func Run(w *Worker) {

}

func NewWorker(ctx context.Context, rpcServer *messenger.RpcServer, config *Config) *Worker {
	ctx, cancel := context.WithCancel(ctx)

	w := &Worker{
		ctx:      ctx,
		cancel:   cancel,
		JobQueue: make(chan service.Job, 100),
		config:   config,
	}

	RegisterWorkerServer(rpcServer.Server, w)

	return w

}
