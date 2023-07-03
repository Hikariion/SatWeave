package worker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io"
	"os"
	"satweave/messenger"
	"satweave/shared/service"
	"satweave/utils/logger"
	"strings"
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

// ExecuteJob 执行任务
func (w *Worker) ExecuteJob(ctx context.Context, job *service.Job) error {
	// 执行任务
	logger.Infof("begin to execute job: %v", job)

	imageName := job.ImageName // 任务类型（镜像名）
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Errorf("Failed to create docker client", err)
		return err
	}
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   strings.Split(job.Command, " "),
		// 伪终端
		Tty: true,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			"/Users/bytedance/code/personal/SatWeave/client/offload/satweave-data/attachment:/usr/src/app/data/images",
			"/Users/bytedance/code/personal/SatWeave/client/offload/satweave-data/output:/usr/src/app/runs/detect/labels",
		},
		PortBindings: nat.PortMap{},
	}

	// 创建容器
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		logger.Errorf("failed to create container", err)
		return err
	}

	// 启动容器
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		logger.Errorf("Failed to start container", err)
		return err
	}

	// Wait for container to finish
	waitCh, _ := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	//if errC != nil {
	//	logger.Errorf("Error waiting for container: %s", errC)
	//}

	// This will block until the container exits
	res := <-waitCh

	if res.Error != nil {
		logger.Errorf("Error from exited container: %s", res.Error.Message)
	}

	// 删除容器，但是不删除相关联的卷
	if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{RemoveVolumes: false}); err != nil {
		logger.Errorf("failed to remove container %v", err)
	}

	// 将任务结果返回给客户端
	return nil
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
			// TODO(qiu): 如果返回错误，要告诉客户端任务执行失败
			go w.ExecuteJob(w.ctx, job)
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

func (w *Worker) Stop() {
	w.cancel()
}
