package task_manager

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"satweave/utils/logger"
)

type TaskManager struct {
	UnimplementedTaskManagerServiceServer
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
