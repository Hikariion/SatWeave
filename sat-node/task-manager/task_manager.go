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
	nextWorkerId    uint64
}

func (t *TaskManager) GetNextWorkerId() uint64 {
	return atomic.AddUint64(&t.nextWorkerId, 1) - 1
}

func (t *TaskManager) newSelfDescription(satelliteName string, slotNum uint64, host string, port uint64) *common.TaskManagerDescription {
	return &common.TaskManagerDescription{
		SatelliteName: satelliteName,
		SlotNumber:    slotNum,
		HostPort: &common.HostPort{
			Host: host,
			Port: port,
		},
	}
}

func (t *TaskManager) PushRecord(_ context.Context, request *task_manager.PushRecordRequest) (*common.NilResponse, error) {
	workerId := request.WorkerId
	// TODO(qiu): 增加转发模式
	worker, err := t.slotTable.getWorkerByWorkerId(workerId)
	if err != nil {
		logger.Errorf("satellite: %s get worker id %v failed: %v", t.selfDescription.SatelliteName, workerId, err)
		return nil, status.Errorf(codes.Internal, "get worker failed: %v", err)
	}
	err = worker.PushRecord(request.Record, request.FromSubtask, request.PartitionIdx)
	if err != nil {
		logger.Errorf("task manager id: %v push record to worker id %v failed: %v", t.selfDescription.SatelliteName, workerId, err)
		return nil, status.Errorf(codes.Internal, "push record failed: %v", err)
	}
	return &common.NilResponse{}, nil
}

func (t *TaskManager) DeployTask(_ context.Context, request *task_manager.DeployTaskRequest) (*common.NilResponse, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.selfDescription.SlotNumber <= 0 {
		logger.Errorf("task manager id: %v slot number is %v, not enough", t.selfDescription.SatelliteName)
		return nil, errno.SlotCapacityNotEnough
	}

	err := t.slotTable.deployExecuteTask(request.JobId, request.ExecTask, request.PathNodes, request.YamlBytes)
	if err != nil {
		logger.Errorf("deploy execute task %v failed, err: %v", request.ExecTask, err)
	}

	t.selfDescription.SlotNumber--
	return &common.NilResponse{}, nil
}

func (t *TaskManager) StartTask(_ context.Context, request *task_manager.StartTaskRequest) (*common.NilResponse, error) {
	subtaskName := request.SubtaskName
	logger.Infof("task manager name %v begin to start %v", t.selfDescription.SatelliteName, subtaskName)
	slot := t.slotTable.getSlot(subtaskName)
	go slot.start()
	logger.Infof("return response")
	return &common.NilResponse{}, nil
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
	_, err = c.RegisterTaskManager(context.Background(), &sun.RegisterTaskManagerRequest{
		TaskManagerDesc: t.selfDescription,
	})
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
	logger.Infof("register to cloud success")
}

func (t *TaskManager) RequestSlot(_ context.Context, request *task_manager.RequiredSlotRequest) (*task_manager.RequiredSlotResponse, error) {
	//if request.RequestSlotNum > t.selfDescription.SlotNumber {
	//	logger.Errorf("request slot num is larger than free worker num")
	//	return &task_manager.RequiredSlotResponse{
	//		AvailableWorkers: nil,
	//		Status: &common.Status{
	//			ErrCode: errno.CodeRequestSlotFail,
	//		},
	//	}, errno.RequestSlotFail
	//}
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

func NewTaskManager(ctx context.Context, config *Config, satelliteName string, server *messenger.RpcServer,
	slotNum uint64, host string, port uint64) *TaskManager {
	taskManagerCtx, cancelFunc := context.WithCancel(ctx)

	taskManager := &TaskManager{
		ctx:          taskManagerCtx,
		config:       config,
		cancelFunc:   cancelFunc,
		mutex:        &sync.Mutex{},
		nextWorkerId: 0,
	}

	taskManager.slotTable = NewSlotTable(satelliteName, slotNum, taskManager.config.CloudAddr, taskManager.config.CloudPort)
	taskManager.selfDescription = taskManager.newSelfDescription(satelliteName, slotNum, host, port)

	task_manager.RegisterTaskManagerServiceServer(server, taskManager)
	return taskManager
}
