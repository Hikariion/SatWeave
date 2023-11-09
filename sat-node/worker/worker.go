package worker

/*
	slot 中实际执行的 subtask(worker)
*/

import (
	"satweave/messenger/common"
	"satweave/sat-node/operators"
	"satweave/utils/logger"
	"sync"
)

type workerState uint64

const (
	workerIdle workerState = iota
	workerBusy
)

type Worker struct {
	// 算子名
	clsName         string
	inputEndpoints  []*common.InputEndpoints
	outputEndpoints []*common.OutputEndpoints
	subTaskName     string
	partitionIdx    uint64

	inputReceiver   *InputReceiver
	outputDispenser *OutputDispenser

	// 存储算子的map
	functionMap map[string]func(data string)
	// 表示该worker是否可用，true 表示可用 false 表示不可用
	state workerState
	mu    sync.Mutex
}

func (w *Worker) initForStartService() {

}

func (w *Worker) isSourceOp() bool {
	// TODO(qiu): need to complete
	return false
}

func (w *Worker) isSinkOp() bool {
	// TODO(qiu): needo to complete
	return false
}
func (w *Worker) IsAvailable() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state == workerIdle
}

func (w *Worker) ComputeCore(clsName string, inputChannel, outputChannel chan *common.Record) {

}

func (w *Worker) innerComputeCore(clsName string, inputChannel, outputChannel chan *common.Record) error {
	// 具体执行逻辑
	var taskInstance operators.OperatorBase
	for {
		dataID := "dataID" // TODO(进程安全) gen
		timestamp := utils.
	}
	return nil
}

func (w *Worker) initInputReceiver(input)  {

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

func (w *Worker) PushRecord(record *common.Record, fromSubTask string, partitionIdx uint64) error {
	preSubTask := fromSubTask
	logger.Infof("Recv data(from=%s): %v", preSubTask, record)
	w.inputReceiver.RecvData(partitionIdx, record)
	return nil
}

func NewWorker() *Worker {
	worker := &Worker{}
	// TODO(qiu)：初始化 functionMap
	worker.state = workerIdle
	return worker
}
