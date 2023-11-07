package task_manager

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/worker"
	"satweave/shared/task-manager"
	"satweave/utils/errno"
	"satweave/utils/logger"
	"sync"
)

type TaskManager struct {
	task_manager.UnimplementedTaskManagerServiceServer
	ctx             context.Context
	config          *Config
	workers         []*worker.Worker
	mutex           sync.Mutex
	selfDescription *common.TaskManagerDescription

	cancelFunc context.CancelFunc
}

func (t *TaskManager) initWorkers() error {
	t.workers = make([]*worker.Worker, t.config.SlotNum)
	for i := 0; i < t.config.SlotNum; i++ {
		t.workers[i] = worker.NewWorker()
	}
	return nil
}

func (t *TaskManager) newSelfDescription(raftId uint64, slotNum uint64, host string, port uint64) *common.TaskManagerDescription {
	return &common.TaskManagerDescription{
		RaftId:     raftId,
		SlotNumber: slotNum,
		Host:       host,
		Port:       port,
	}
}

func (t *TaskManager) PushRecord(_ context.Context, request *task_manager.PushRecordRequest) (*common.NilResponse, error) {
	workerId := request.WorkerId
	err := t.workers[workerId].PushRecord(request.Record, request.FromSubtask, request.PartitionIdx)
	if err != nil {
		logger.Errorf("task manager id: %v push record to worker id %v failed: %v", t.selfDescription.RaftId, workerId, err)
		return nil, status.Errorf(codes.Internal, "push record failed: %v", err)
	}
	return nil, nil
}

func (t *TaskManager) DeployTask(_ context.Context, request *common.ExecuteTask) (*common.NilResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method DeployTask not implemented")
}

func (t *TaskManager) registerToCloud() error {
	if t.config.CloudAddr == "" {
		logger.Errorf("cloud addr is empty")
		return errno.ConnectSunFail
	}

	conn, err := grpc.Dial(t.config.SunAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Errorf("connect cloud failed: %v", err)
		return err
	}

	c := sun.NewSunClient(conn)
	_, err = c.RegisterTaskManager(context.Background(), t.selfDescription)
	if err != nil {
		logger.Errorf("register to cloud failed: %v", err)
		return err
	}

	return nil
}

func (t *TaskManager) Run() {
	// 向 Cloud 注册
	err := t.registerToCloud()
	if err != nil {
		logger.Errorf("register to cloud failed: %v", err)
	}
}

func (t *TaskManager) GetFreeWorkerIdList() []uint64 {
	var freeWorkerIdList []uint64
	for i := 0; i < len(t.workers); i++ {
		if t.workers[i].IsAvailable() == true {
			freeWorkerIdList = append(freeWorkerIdList, uint64(i))
		}
	}
	return freeWorkerIdList
}

func (t *TaskManager) RequestSlot(_ context.Context, request *task_manager.RequiredSlotRequest) (*task_manager.RequiredSlotResponse, error) {
	freeWorkerIdList := t.GetFreeWorkerIdList()
	if request.RequestSlotNum > uint64(len(freeWorkerIdList)) {
		logger.Errorf("request slot num is larger than free worker num")
		return &task_manager.RequiredSlotResponse{
			AvailableWorkers: nil,
			Status: &common.Status{
				ErrCode: errno.CodeRequestSlotFail,
			},
		}, errno.RequestSlotFail
	}

	// 返回可用的workerId
	assignedWorkerIdList := freeWorkerIdList[:request.RequestSlotNum]
	return &task_manager.RequiredSlotResponse{
		AvailableWorkers: assignedWorkerIdList,
	}, nil
}

func (t *TaskManager) Stop() {
	t.cancelFunc()
}

func NewTaskManager(ctx context.Context, config *Config, raftID uint64, server *messenger.RpcServer,
	slotNum uint64, host string, port uint64) *TaskManager {
	taskManagerCtx, cancelFunc := context.WithCancel(ctx)

	taskManager := &TaskManager{
		ctx:        taskManagerCtx,
		config:     config,
		cancelFunc: cancelFunc,
	}

	taskManager.selfDescription = taskManager.newSelfDescription(raftID, slotNum, host, port)
	taskManager.workers = make([]*worker.Worker, slotNum)
	for i := 0; i < int(taskManager.selfDescription.SlotNumber); i++ {
		taskManager.workers[i] = worker.NewWorker()
	}

	task_manager.RegisterTaskManagerServiceServer(server, taskManager)
	return taskManager
}
