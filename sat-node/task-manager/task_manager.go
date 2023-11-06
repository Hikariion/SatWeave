package task_manager

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

//func (t *TaskManager) ProcessOperator(ctx context.Context, request *OperatorRequest) (*common.NilResponse, error) {
//
//	return &common.NilResponse{}, nil
//}

//func (t *TaskManager) AskAvailableWorkers(context.Context, *common.NilRequest) (*AvailableWorkersResponse, error) {
//	t.mutex.Lock()
//	defer t.mutex.Unlock()
//	unusedWorks := make([]int64, 0)
//
//	for i := 0; i < t.config.SlotNum; i++ {
//		if t.workers[i].Available() {
//			unusedWorks = append(unusedWorks, int64(i))
//		}
//	}
//
//	return &AvailableWorkersResponse{
//		Workers: unusedWorks,
//	}, nil
//}
//
//type TaskYaml struct {
//	Tasks     []*Task `yaml:"tasks"`
//	TaskFiles string  `yaml:"task_files"`
//}

//func GetLogicalTaskFromYaml(yamlPath string) (*TaskYaml, error) {
//	yamlFile, err := ioutil.ReadFile(yamlPath)
//	if err != nil {
//		logger.Errorf("Error reading YAML file: %s\n", err)
//		return nil, err
//	}
//
//	taskYaml := new(TaskYaml)
//	err = yaml.Unmarshal(yamlFile, taskYaml)
//	if err != nil {
//		logger.Errorf("Error parsing YAML file: %s\n", err)
//		return nil, err
//	}
//	return nil, nil
//}

func (t *TaskManager) newSelfDescription(raftId uint64, slotNum uint64, host string, port uint64) *common.TaskManagerDescription {
	return &common.TaskManagerDescription{
		RaftId:     raftId,
		SlotNumber: slotNum,
		Host:       host,
		Port:       port,
	}
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
		taskManager.workers[i] = nil
	}

	task_manager.RegisterTaskManagerServiceServer(server, taskManager)
	return taskManager
}
