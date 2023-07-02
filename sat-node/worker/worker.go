package worker

import (
	"context"
	"io"
	"os"
	"satweave/messenger"
	"satweave/shared/service"
	"satweave/utils/logger"
	"sync"
)

// 该模块用于执行任务

type Worker struct {
	UnimplementedWorkerServer

	// TODO(qiutb): 要不要加上worker的id？
	ctx    context.Context
	cancel context.CancelFunc

	// 任务队列, 用于接收任务
	JobQueue chan *service.Job
	// 配置
	config *Config
	mutex  sync.Mutex
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

// SubmitJob 客户端通过这个方法提交任务
func (w *Worker) SubmitJob(ctx context.Context, request *SubmitJobRequest) (*SubmitJobReply, error) {
	job := request.Job
	//job.Status = service.JobStatus_ASSIGNED
	w.JobQueue <- job
	return &SubmitJobReply{
		Success: true,
	}, nil
}

// 执行任务
func (w *Worker) executeJob(job *service.Job) {
	// 从任务队列里取任务
	// 执行任务
	// 将任务结果返回给客户端
}

func (w *Worker) Run() {
	// 监控其他节点的状态，做迁移决策
	// go XXXXXXX
	for {
		select {
		case <-w.ctx.Done():
			w.cleanup()
			return
		// 从任务队列里取任务
		case job := <-w.JobQueue:
			w.executeJob(job)
		}
	}
}

func (w *Worker) cleanup() {
	logger.Infof("Worker cleanup")
}

func NewWorker(ctx context.Context, rpcServer *messenger.RpcServer, config *Config) *Worker {
	ctx, cancel := context.WithCancel(ctx)

	w := &Worker{
		ctx:      ctx,
		cancel:   cancel,
		JobQueue: make(chan *service.Job, 100),
		config:   config,
	}

	RegisterWorkerServer(rpcServer.Server, w)

	return w

}
