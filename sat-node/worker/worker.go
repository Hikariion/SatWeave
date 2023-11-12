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

type Worker struct {
	// raftID
	raftId uint64
	// 算子名
	clsName string
	// 用 UUID 唯一标识这个 subtask
	subtaskId       string
	workerId        uint64
	inputEndpoints  []*common.InputEndpoints
	outputEndpoints []*common.OutputEndpoints
	subTaskName     string
	partitionIdx    int64

	inputReceiver   *InputReceiver
	outputDispenser *OutputDispenser

	cls operators.OperatorBase

	mu sync.Mutex
}

func (w *Worker) startComputeOnStandletonProcess(inputChannel chan *common.Record, outputChannel chan *common.Record) {
	w.ComputeCore(w.clsName, inputChannel, outputChannel)
}

func (w *Worker) initForStartService() {
	var inputChannel chan *common.Record
	var outputChannel chan *common.Record
	if !w.isSourceOp() {
		inputChannel = w.initInputReceiver(w.inputEndpoints)
	}
	if !w.isSinkOp() {
		outputChannel = w.initOutputDispenser(w.outputEndpoints)
	}
	w.startComputeOnStandletonProcess(inputChannel, outputChannel)
}

// 判断是否为 source 算子
func (w *Worker) isSourceOp() bool {
	return w.cls.IsSourceOp()
}

// 判断是否为 sink 算子
func (w *Worker) isSinkOp() bool {
	return w.cls.IsSinkOp()
}

// 判断是否为 key 算子
func (w *Worker) isKeyOp() bool {
	return w.cls.IsKeyByOp()
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

func (w *Worker) ComputeCore(clsName string, inputChannel, outputChannel chan *common.Record) {
	// TODO(qiu):SourceOp 中通过 StopIteration 异常（迭代器终止）来表示
	// 用 error 代替？
	go func() {
		err := w.innerComputeCore(clsName, inputChannel, outputChannel)
		if err != nil {
			logger.Errorf("ComputeCore error: %v", err)
			// TODO(qiu): 添加错误处理
		}
	}()
}

func (w *Worker) innerComputeCore(clsName string, inputChannel, outputChannel chan *common.Record) error {
	// 具体执行逻辑
	cls := operators.FactoryMap[clsName]()
	w.cls = cls
	// 判断是否为 source 算子
	isSourceOp := w.isSourceOp()
	// 判断是否为 sink 算子
	isSinkOp := w.isSinkOp()
	// 判断是否为 key 算子
	isKeyOp := w.isKeyOp()

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
		} else {
			dataId = "data_id"
			timestamp = timestampUtil.GetTimeStamp()
		}

		if isKeyOp {
			outputData = inputData
			// TODO(qiu): 把 Compute 的返回值都改成 bytes
			key, err := taskInstance.Compute(inputData)
			if err != nil {
				logger.Errorf("Compute error: %v", err)
				// return err
			}
			partitionKey = int(key)
		} else {
			data, err := taskInstance.Compute(inputData)
			outputData = data
		}

		if !isSinkOp {
			// TODO(qiu) 将 output_data 转成 record
			output := &common.Record{}
			outputChannel <- output
		}
	}
}

func (w *Worker) initOutputDispenser(outputEndpoints []*common.OutputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	outputChannel := make(chan *common.Record, 1000)
	// TODO(qiu): 研究一下 PartitionIdx 的作用
	w.outputDispenser = NewOutputDispenser(outputChannel, outputEndpoints, w.subTaskName, w.partitionIdx)
	return outputChannel

}

func (w *Worker) PushRecord(record *common.Record, fromSubTask string, partitionIdx int64) error {
	preSubTask := fromSubTask
	logger.Infof("Recv data(from=%s): %v", preSubTask, record)
	w.inputReceiver.RecvData(partitionIdx, record)
	return nil
}

// Run 启动 Worker
func (w *Worker) Run() {
	// 启动 Receiver 和 Dispenser

	// 启动 worker 核心进程
}

func NewWorker(raftId uint64, executeTask *common.ExecuteTask) *Worker {
	worker := &Worker{
		raftId:          raftId,
		subtaskId:       uuid.New().String(),
		clsName:         executeTask.ClsName,
		inputEndpoints:  executeTask.InputEndpoints,
		outputEndpoints: executeTask.OutputEndpoints,
		subTaskName:     executeTask.SubtaskName,
		partitionIdx:    executeTask.PartitionIdx,
		workerId:        executeTask.WorkerId,

		inputReceiver:   nil,
		outputDispenser: nil,
	}
	return worker
}
