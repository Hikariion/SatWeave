package worker

/*
	slot 中实际执行的 subtask(worker)
*/

import (
	"context"
	"github.com/google/uuid"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/operators"
	common2 "satweave/utils/common"
	"satweave/utils/errno"
	"satweave/utils/logger"
	"sync"
)

const (
	lastFinishDataId uint64 = 5000 + iota
	checkpointDataId
)

type Worker struct {
	// satellite name
	satelliteName string
	// 算子名
	clsName string
	jobId   string
	// 用 UUID 唯一标识这个 subtask
	subtaskId       string
	workerId        uint64
	inputEndpoints  []*common.InputEndpoints
	outputEndpoints []*common.OutputEndpoints
	SubTaskName     string
	partitionIdx    int64

	// 数据接收器
	InputReceiver *InputReceiver
	// 数据分发器
	OutputDispenser *OutputDispenser

	cls operators.OperatorBase

	// 如果是 Source 算子，这个Channel用于接收Event
	// 否则，这个 Channel 用于接收数据
	InputChannel  chan *common.Record
	OutputChannel chan *common.Record

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.Mutex

	CheckpointDir string

	// job manager endpoint
	jobManagerHost string
	jobManagerPort uint64

	// 初始化时的 state
	state *common.File
}

// ----------------------------------- init for start subtask -----------------------------------

// InitForStartService 用于初始化 Worker
func (w *Worker) InitForStartService() {
	if !w.isSourceOp() {
		w.InputChannel = w.initInputReceiver(w.inputEndpoints)
	} else {
		// 如果是 SourceOp
		//这个Channel 用来接收 checkpoint 等 event
		w.InputChannel = make(chan *common.Record, 1000)
	}
	if !w.isSinkOp() {
		w.OutputChannel = w.initOutputDispenser(w.outputEndpoints)
	}
}

func (w *Worker) initInputReceiver(inputEndpoints []*common.InputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	inputChannel := make(chan *common.Record, 10000)
	w.InputReceiver = NewInputReceiver(w.ctx, inputChannel, inputEndpoints)
	return inputChannel
}

func (w *Worker) initOutputDispenser(outputEndpoints []*common.OutputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	outputChannel := make(chan *common.Record, 10000)
	// TODO(qiu): 研究一下 PartitionIdx 的作用
	w.OutputDispenser = NewOutputDispenser(w.ctx, outputChannel, outputEndpoints, w.SubTaskName, w.partitionIdx)
	return outputChannel
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

// --------------------------- Compute Core ----------------------------

func (w *Worker) ComputeCore() error {
	// ------------------------ 具体执行逻辑 ------------------------

	// 判断是否为 source 算子
	isSourceOp := w.isSourceOp()
	// 判断是否为 sink 算子
	isSinkOp := w.isSinkOp()
	// 判断是否为 key 算子
	isKeyOp := w.isKeyOp()

	taskInstance := w.cls
	taskInstance.SetName(w.SubTaskName)

	initMap := make(map[string]interface{})
	initMap["InputChannel"] = w.InputChannel

	taskInstance.RestoreFromCheckpoint(w.jobManagerHost, w.SubTaskName, w.jobManagerPort)
	taskInstance.Init(initMap)

	for {
		flag := false
		select {
		case <-w.ctx.Done():
			logger.Infof("finish cls %v", w.clsName)
			return nil
		default:
			var partitionKey int64 = 0
			var outputData []byte = nil
			var err error = nil

			dataType, inputData := w.getInputData(isSourceOp)

			// TODO 1.31 都还没做 error 的处理
			if dataType == common.DataType_BINARY {
				outputData, partitionKey, err = w.BinaryDataProcess(taskInstance, isKeyOp, isSourceOp, inputData)
				if err != nil {
					logger.Errorf("Process BinaryData Err: %v, err")
					return err
				}
			} else if dataType == common.DataType_CHECKPOINT {
				data := w.checkpointEventProcess(isSinkOp)
				if data != nil {
					conn, err := messenger.GetRpcConn(w.jobManagerHost, w.jobManagerPort)
					if err != nil {
						logger.Errorf("GetRpcConn Err: %v", err)
						return err
					}
					client := sun.NewSunClient(conn)
					result, err := client.SaveSnapShot(context.Background(), &sun.SaveSnapShotRequest{
						FilePath: w.SubTaskName,
						State:    data,
					})
					if err != nil {
						logger.Errorf("SaveSnapShot Err: %v", err)
						return err
					}
					logger.Infof("SaveSnapShot result: %v", result)
				}

			} else if dataType == common.DataType_FINISH {
				_ = w.finishEventProcess(taskInstance, isSinkOp, inputData)
				logger.Infof("%s finished successfully!", w.SubTaskName)
				flag = true
			} else {
				logger.Errorf("Failed unknown data type")
				return errno.UnknownDataType
			}

			w.pushOutputData(isSinkOp, outputData, dataType, partitionKey)
		}
		if flag {
			break
		}
	}
	w.cancel()
	return nil
}

/* record.data
如果record type 是 pickle，则 data 是 raw bytes
如果record type 是 checkpoint，则 data 需要转成 checkpoint 结构体
如果record type 是 finish，则 data 需要转成 finish 结构体
/
*/

// -------------------- get input --------------------

// 获取算子的输入数据，返回数据类型、内容、数据ID
func (w *Worker) getInputData(isSourceOp bool) (common.DataType, []byte) {
	// 不是 SourceOp, 从 inputChannel 正常获取数据
	if !isSourceOp {
		record := <-w.InputChannel
		return record.DataType, record.Data
	}

	// 是SourceOp, 从 inputChannel 里获取 event
	record := <-w.InputChannel
	return record.DataType, record.Data
}

// -------------------- process data --------------------

// BinaryDataProcess 处理序列化的二进制数据
func (w *Worker) BinaryDataProcess(taskInstance operators.OperatorBase, isKeyOp bool, isSourceOp bool,
	inputData []byte) ([]byte, int64, error) {
	var outputData []byte = nil
	var partitionKey int64 = -1
	var err error = nil

	// KeyOp 可以不管了，不会有这个类型
	if isKeyOp {
		// 如果是 Key Operator，outputData = inputData
		// 获取 partitionKey
		outputData = inputData
		// TODO: 把 partitionKey 算子的 Compute 函数中 int64 转成 []byte 的操作统一
		partitionKeyBytes, err := taskInstance.Compute(inputData)
		if err != nil {
			logger.Errorf("Compute error: %v", err)
			return nil, 0, err
		}
		partitionKey = common2.BytesToInt64(partitionKeyBytes)
	} else if isSourceOp {
		// SourceOp 在这里处理
		outputData = inputData
	} else {
		// 不是 partition key operator以及SourceOp, 计算 outputData
		outputData, err = taskInstance.Compute(inputData)
	}
	return outputData, partitionKey, err
}

func (w *Worker) finishEventProcess(taskInstance operators.OperatorBase, isSinkOp bool, inputData []byte) error {
	if isSinkOp {
		return nil
	}
	w.PushFinishEventToOutputChannel()
	return nil
}

// -------------------- push output --------------------
func (w *Worker) pushOutputData(isSinkOp bool, outputData []byte, dataType common.DataType, partitionKey int64) {
	if isSinkOp {
		return
	}
	record := &common.Record{
		DataType:     dataType,
		Data:         outputData,
		PartitionKey: partitionKey,
	}
	w.OutputChannel <- record
}

// PushRecord 用于处理 OutputDispatcher 的 Rpc 请求
// 将数据写入 Worker 的 InputReceiver
func (w *Worker) PushRecord(record *common.Record, fromSubTask string, partitionIdx int64) error {
	preSubTask := fromSubTask
	logger.Infof("Recv data(from=%s): %v", preSubTask, record)
	w.InputReceiver.RecvData(partitionIdx, record)
	return nil
}

func (w *Worker) PushEventRecordToOutputChannel(dataType common.DataType,
	data []byte) {
	if w.isSinkOp() {
		return
	}
	record := &common.Record{
		DataType:     dataType,
		Data:         data,
		PartitionKey: -1,
	}
	w.OutputChannel <- record
}

func (w *Worker) PushFinishEventToOutputChannel() {
	w.PushEventRecordToOutputChannel(common.DataType_FINISH, nil)
}

// --------------------------- checkpoint ----------------------------
func (w *Worker) checkpointEventProcess(isSinkOp bool) []byte {
	if isSinkOp {
		return nil
	}
	data := w.cls.Checkpoint()

	w.PushCheckpointEventToOutputChannel(w.OutputChannel)

	return data
}

func (w *Worker) PushCheckpointEventToOutputChannel(outputChannel chan *common.Record) {
	w.PushEventRecordToOutputChannel(common.DataType_CHECKPOINT, nil)
}

// --------------------------- Run ----------------------------

// Run 启动 Worker
func (w *Worker) Run() {
	// TODO(qiu)：都还需要添加错误处理
	// 启动 Receiver 和 Dispenser
	if w.InputReceiver != nil {
		w.InputReceiver.RunAllPartitionReceiver()
		logger.Infof("subtask %v start input receiver finish...", w.SubTaskName)
	}

	if w.OutputDispenser != nil {
		w.OutputDispenser.Run()
		logger.Infof("subtask %v start output dispenser finish...", w.SubTaskName)
	}

	// 启动 worker 核心进程
	err := w.ComputeCore()
	if err != nil {
		logger.Errorf("ComputeCore Err: %v", err)
	}
	//time.Sleep(60 * time.Second)
	logger.Infof("%s start core compute process success...", w.SubTaskName)
}

func (w *Worker) Stop() {
	w.cancel()
}

func NewWorker(satelliteName string, executeTask *common.ExecuteTask, jobManagerHost string, jobManagerPort uint64, jobId string,
) *Worker {
	// TODO(qiu): 这个 ctx 是否可以继承 task manager
	workerCtx, cancel := context.WithCancel(context.Background())
	worker := &Worker{
		satelliteName:   satelliteName,
		subtaskId:       uuid.New().String(),
		clsName:         executeTask.ClsName,
		inputEndpoints:  executeTask.InputEndpoints,
		outputEndpoints: executeTask.OutputEndpoints,
		SubTaskName:     executeTask.SubtaskName,
		partitionIdx:    executeTask.PartitionIdx,
		workerId:        executeTask.WorkerId,

		InputReceiver:   nil,
		OutputDispenser: nil,

		ctx:    workerCtx,
		cancel: cancel,

		jobManagerHost: jobManagerHost,
		jobManagerPort: jobManagerPort,
		jobId:          jobId,
	}

	// init op
	cls := operators.FactoryMap[worker.clsName]()
	worker.cls = cls

	// worker 初始化
	worker.InitForStartService()

	return worker
}
