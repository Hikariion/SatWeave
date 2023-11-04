package task_manager

import (
	"context"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"satweave/messenger/common"
	"satweave/sat-node/watcher"
	"satweave/utils/logger"
	"sync"
)

type TaskManager struct {
	UnimplementedTaskManagerServiceServer
	config  *Config
	watcher *watcher.Watcher
	workers []*Worker

	mutex sync.Mutex
}

func (t *TaskManager) initWorkers() error {
	t.workers = make([]*Worker, t.config.SlotNum)
	for i := 0; i < t.config.SlotNum; i++ {
		t.workers[i] = NewWorker()
	}
	return nil
}

func (t *TaskManager) ProcessOperator(ctx context.Context, request *OperatorRequest) (*common.NilResponse, error) {

	return &common.NilResponse{}, nil
}

func (t *TaskManager) AskAvailableWorkers(context.Context, *common.NilRequest) (*AvailableWorkersResponse, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	unusedWorks := make([]int64, 0)

	for i := 0; i < t.config.SlotNum; i++ {
		if t.workers[i].Available() {
			unusedWorks = append(unusedWorks, int64(i))
		}
	}

	return &AvailableWorkersResponse{
		Workers: unusedWorks,
	}, nil
}

type TaskYaml struct {
	Tasks     []*Task `yaml:"tasks"`
	TaskFiles string  `yaml:"task_files"`
}

func GetLogicalTaskFromYaml(yamlPath string) (*TaskYaml, error) {
	yamlFile, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		logger.Errorf("Error reading YAML file: %s\n", err)
		return nil, err
	}

	taskYaml := new(TaskYaml)
	err = yaml.Unmarshal(yamlFile, taskYaml)
	if err != nil {
		logger.Errorf("Error parsing YAML file: %s\n", err)
		return nil, err
	}
	return nil, nil
}
