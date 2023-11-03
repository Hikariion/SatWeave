package task_manager

import "satweave/sat-node/task"

type Worker struct {
	functionMap map[string]func(data string)
}

func NewWorker() *Worker {
	worker := &Worker{}

	// 初始化 functionMap
	worker.functionMap["sink"] = task.Sink
	worker.functionMap["source"] = task.Source

	return &Worker{
		functionMap: make(map[string]func(data string)),
	}
}
