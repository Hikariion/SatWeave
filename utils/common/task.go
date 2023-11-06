package common

import (
	"gopkg.in/yaml.v2"
	"os"
	"satweave/messenger/common"
	"satweave/utils/logger"
)

// UserTaskDefine 用户定义的逻辑Task，用于解析yaml
type UserTaskDefine struct {
	Cls        string   `yaml:"cls"`
	Currency   int64    `yaml:"currency"`
	InputTasks []string `yaml:"input_tasks"`
	Locate     uint64   `yaml:"locate"`
}

type UserTaskWrapper struct {
	Tasks []UserTaskDefine `yaml:"tasks"`
}

func ReadUserDefinedTasks(filePath string) (*UserTaskWrapper, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Errorf("ReadUserDefinedTasks() failed: %v", err)
		return nil, err
	}

	var tasksWrapper UserTaskWrapper

	err = yaml.Unmarshal(data, &tasksWrapper)
	if err != nil {
		logger.Errorf("ReadUserDefinedTasks() failed: %v", err)
		return nil, err
	}
	return &tasksWrapper, nil
}

func ConvertUserTaskWrapperToLogicTasks(userTaskWrapper *UserTaskWrapper) ([]*common.Task, error) {
	logicalTasks := make([]*common.Task, 0)
	for _, userTask := range userTaskWrapper.Tasks {
		logicalTask := &common.Task{
			ClsName:    userTask.Cls,
			Currency:   userTask.Currency,
			InputTasks: userTask.InputTasks,
			Locate:     userTask.Locate,
		}
		logicalTasks = append(logicalTasks, logicalTask)
	}
	return logicalTasks, nil
}
