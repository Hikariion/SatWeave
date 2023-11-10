package worker

/*
	slot 中实际执行的 subtask(worker)
*/

import (
	"github.com/google/uuid"
	"satweave/messenger/common"
	"satweave/sat-node/operators"
	"satweave/utils/logger"
	timestampUtil "satweave/utils/timestamp"
	"sync"
)

type workerState uint64

const (
	workerIdle workerState = iota
	workerBusy
)

type Worker struct {
	// 算子名
	clsName string
	// 用 UUID 唯一标识这个 worker
	workerId        string
	inputEndpoints  []*common.InputEndpoints
	outputEndpoints []*common.OutputEndpoints
	subTaskName     string
	partitionIdx    uint64

	inputReceiver   *InputReceiver
	outputDispenser *OutputDispenser

	// 表示该worker是否可用，true 表示可用 false 表示不可用
	state workerState
	mu    sync.Mutex
}

func (w *Worker) initForStartService() {

}

// 判断是否为 source 算子
func (w *Worker) isSourceOp(cls operators.OperatorBase) bool {
	if _, ok := interface{}(cls).(operators.SourceOperatorBase); ok {
		return true
	}
	return false
}

// 判断是否为 sink 算子
func (w *Worker) isSinkOp(cls operators.OperatorBase) bool {
	if _, ok := interface{}(cls).(operators.SinkOperatorBase); ok {
		return true
	}
	return false
}

// 判断是否为 key 算子
func (w *Worker) isKeyOp(cls operators.OperatorBase) bool {
	if _, ok := interface{}(cls).(operators.KeyOperatorBase); ok {
		return true
	}
	return false
}

func (w *Worker) pushFinishRecordToOutPutChannel(outputChannel chan *common.Record) {
	record := &common.Record{
		DataId:   "last_finish_data_id",
		DataType: common.DataType_FINISH,
		// TODO(qiu): 没写完，只有架子
	}
	outputChannel <- record
}

func (w *Worker) initInputReceiver(inputEndpoints []*common.InputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	inputChannel := make(chan *common.Record, 1000)
	w.inputReceiver = NewInputReceiver(inputChannel, inputEndpoints)
	return inputChannel
}

func (w *Worker) IsAvailable() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state == workerIdle
}

func (w *Worker) ComputeCore(clsName string, inputChannel, outputChannel chan *common.Record) {
	// TODO(qiu):SourceOp 中通过 StopIteration 异常（迭代器终止）来表示
	// 用 error 代替？
}

func (w *Worker) innerComputeCore(clsName string, inputChannel, outputChannel chan *common.Record) error {
	// 具体执行逻辑
	cls := operators.FactoryMap[clsName]()
	// 判断是否为 source 算子
	isSourceOp := w.isSourceOp(cls)
	// 判断是否为 sink 算子
	isSinkOp := w.isSinkOp(cls)
	// 判断是否为 key 算子
	isKeyOp := w.isKeyOp(cls)

	taskInstance := cls
	cls.SetName(w.subTaskName)
	// TODO(qiu): cls Init
	// TODO(qiu): succ start event?

	for {
		var dataId string // TODO(进程安全) gen
		var timestamp uint64
		partitionKey := -1
		var inputData []byte
		var outputData []byte
		if !isSourceOp {
			record := <-inputChannel
			// 注意 这里的 data 是 []byte
			inputData = record.Data
			dataId = record.DataId
			timestamp = record.Timestamp

			if record.DataType == common.DataType_FINISH {
				logger.Infof("%v finished successfully!", w.subTaskName)
				if !isSinkOp {
					w.pushFinishRecordToOutPutChannel(outputChannel)
					// 退出循环
					break
				}
			}
		} else {
			dataId = "data_id"
			timestamp = timestampUtil.GetTimeStamp()
		}

		if isKeyOp {
			// TODO(qiu) ???
			// output_data = input_data
			// partitionKey = taskInstance.Compute(inputData)
		} else {
			// outputData = taskInstance.Compute(data)
		}

		if !isSinkOp {
			// TODO(qiu) 将 output_data 转成 record
			output := &common.Record{}
			outputChannel <- output
		}
	}
	return nil
}

func (w *Worker) initOutputDispenser(outputEndpoints []*common.OutputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	outputChannel := make(chan *common.Record, 1000)
	// TODO(qiu): 研究一下 PartitionIdx 的作用
	w.outputDispenser = NewOutputDispenser(outputChannel, outputEndpoints, w.subTaskName, w.partitionIdx)
	return outputChannel

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
	worker := &Worker{
		workerId: uuid.New().String(),
	}
	// TODO(qiu)：初始化 functionMap
	worker.state = workerIdle
	return worker
}
