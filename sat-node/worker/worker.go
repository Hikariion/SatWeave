package worker

import (
	"sync"
)

type WorkerState uint64

const (
	workerIdle WorkerState = iota
	workerBusy
)

type Worker struct {
	// 存储算子的map
	functionMap map[string]func(data string)
	// 表示该worker是否可用，true 表示可用 false 表示不可用
	state WorkerState
	mu    sync.Mutex
}

func (w *Worker) IsAvailable() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state == workerIdle
}

// Set 用于 worker 被调用时的设置
func (w *Worker) Set() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.state = workerBusy
}

func (w *Worker) Free() {
	// TODO(qiu): 清空已有设置
	w.mu.Lock()
	defer w.mu.Unlock()
	w.state = workerIdle
}

func NewWorker() *Worker {
	worker := &Worker{}
	// TODO(qiu)：初始化 functionMap
	worker.state = workerIdle
	return worker
}
