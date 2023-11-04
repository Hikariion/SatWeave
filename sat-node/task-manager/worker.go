package task_manager

import (
	"satweave/sat-node/task"
	"sync"
)

type Worker struct {
	// 存储算子的map
	functionMap map[string]func(data string)
	// 表示该worker是否可用，true 表示可用 false 表示不可用
	state bool
	mu    sync.Mutex
}

func (w *Worker) Available() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state
}

func (w *Worker) Occupy() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.state = false
}

func (w *Worker) Free() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.state = true
}

func NewWorker() *Worker {
	worker := &Worker{}
	// 初始化 functionMap
	worker.functionMap["sink"] = task.Sink
	worker.functionMap["source"] = task.Source
	worker.state = true
	return worker
}
