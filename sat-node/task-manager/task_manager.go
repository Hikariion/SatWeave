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
	"satweave/shared/task-manager"
	"satweave/utils/errno"
	"satweave/utils/logger"
	"sync"
	"sync/atomic"
)

type TaskManager struct {
	task_manager.UnimplementedTaskManagerServiceServer
	ctx    context.Context
	config *Config
	mutex  *sync.Mutex
	// selfDescription 里的 slotNum 是剩余的 slotNum 数量
	selfDescription *common.TaskManagerDescription
	slotTable       *SlotTable
	cancelFunc      context.CancelFunc
	nextWorkerId    atomic.Uint64
}

func (t *TaskManager) GetNextWorkerId() uint64 {
	id := t.nextWorkerId.Load()
	t.nextWorkerId.Add(1)
	return id
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
	// TODO(qiu): 增加转发模式
	worker, err := t.slotTable.getWorkerByWorkerId(workerId)
	if err != nil {
		logger.Errorf("task manager id: %v get worker id %v failed: %v", t.selfDescription.RaftId, workerId, err)
		return nil, status.Errorf(codes.Internal, "get worker failed: %v", err)
	}
	err = worker.PushRecord(request.Record, request.FromSubtask, request.PartitionIdx)
	if err != nil {
		logger.Errorf("task manager id: %v push record to worker id %v failed: %v", t.selfDescription.RaftId, workerId, err)
		return nil, status.Errorf(codes.Internal, "push record failed: %v", err)
	}
	return nil, nil
}

func (t *TaskManager) DeployTask(_ context.Context, executeTask *common.ExecuteTask) (*common.NilResponse, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.selfDescription.SlotNumber <= 0 {
		logger.Errorf("task manager id: %v slot number is %v, not enough", t.selfDescription.RaftId)
		return nil, errno.SlotCapacityNotEnough
	}

	err := t.slotTable.deployExecuteTask(executeTask)
	if err != nil {
		logger.Errorf("deploy execute task %v failed, err: %v", executeTask, err)
	}

	t.selfDescription.SlotNumber--
	return nil, nil
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

func (t *TaskManager) RequestSlot(_ context.Context, request *task_manager.RequiredSlotRequest) (*task_manager.RequiredSlotResponse, error) {
	if request.RequestSlotNum > t.selfDescription.SlotNumber {
		logger.Errorf("request slot num is larger than free worker num")
		return &task_manager.RequiredSlotResponse{
			AvailableWorkers: nil,
			Status: &common.Status{
				ErrCode: errno.CodeRequestSlotFail,
			},
		}, errno.RequestSlotFail
	}

	// 返回可用的workerId
	assignedWorkerIdList := make([]uint64, 0)
	for i := 0; i < int(request.RequestSlotNum); i++ {
		assignedWorkerIdList = append(assignedWorkerIdList, t.GetNextWorkerId())
	}
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
		mutex:      &sync.Mutex{},
	}

	taskManager.slotTable = NewSlotTable(raftID, slotNum)
	taskManager.selfDescription = taskManager.newSelfDescription(raftID, slotNum, host, port)

	task_manager.RegisterTaskManagerServiceServer(server, taskManager)
	return taskManager
}
