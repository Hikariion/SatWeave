package worker

/*
	slot 中实际执行的 subtask(worker)
*/

import (
	"context"
	"github.com/google/uuid"
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

//func (w *Worker) ComputeCore() error {
//	// TODO(qiu):SourceOp 中通过 StopIteration 异常（迭代器终止）来表示
//	// 用 error 代替？
//	errChan := make(chan error, 1)
//	go func() {
//		err := w.innerComputeCore()
//		if err != nil {
//			logger.Errorf("innerComputeCore error: %v", err)
//			errChan <- err
//			return
//		}
//		errChan <- nil
//	}()
//
//	err := <-errChan
//	if err != nil {
//		logger.Errorf("innerComputeCore error: %v", err)
//		if err == errno.JobFinished {
//			logger.Infof("%s finished successfully", w.SubTaskName)
//			w.PushFinishEventToOutputChannel(w.OutputChannel)
//		} else {
//			logger.Errorf("Failed: run %s task failed, err %v", w.SubTaskName, err)
//			return err
//		}
//	}
//	//err := w.innerComputeCore(inputChannel, outputChannel, state)
//	//if err == errno.JobFinished {
//	//	// SourceOp 中通过 JobFinishedError 异常来表示
//	//	// 处理完成。向下游 Operator 发送 Finish Record
//	//	logger.Infof("%s finished successfully!", w.SubTaskName)
//	//	w.PushFinishEventToOutputChannel(outputChannel)
//	//} else {
//	//	logger.Errorf("run %s task failed, err %v", w.SubTaskName, err)
//	//	return err
//	//}
//	return nil
//}

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
	// TODO: 这里的 init 需要完善
	taskInstance.Init(nil)

	// ------------------------ 如果 state 不为空，需要恢复状态 ------------------------
	//if w.state != nil {
	//	// TODO: taskInstance 需要恢复状态
	//	// 恢复状态 ack，其实也可不用，状态还没恢复的话，不会处理数据
	//	// restore from checkpoint
	//	err := taskInstance.RestoreFromCheckpoint(w.state.Content)
	//	if err != nil {
	//		logger.Errorf("%s restore from checkpoint error: %v", w.subtaskId, err)
	//		return errno.RestoreFromCheckpointFail
	//	}
	//}

	for {
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
				outputData, partitionKey, err = w.BinaryDataProcess(taskInstance, isKeyOp, inputData)
				if err != nil {
					logger.Errorf("Process BinaryData Err: %v, err")
					return err
				}
			} else if dataType == common.DataType_CHECKPOINT {
				//_ = w.checkpointEventProcess(taskInstance, isSinkOp, inputData)
				//logger.Infof("%s success save checkpoint state", w.SubTaskName)
				//// 把 inputdata 转成 checkpoint
				//checkpointRecord := &common.Record_Checkpoint{}
				//err := checkpointRecord.Unmarshal(inputData)
				//if err != nil {
				//	logger.Errorf("Fail to unmarchal checkpoint data: %v", err)
				//	return err
				//}
				//cancelJob := checkpointRecord.CancelJob
				//if cancelJob {
				//	logger.Infof("%s success finish job", w.SubTaskName)
				//	break
				//} else {
				//	continue
				//}
			} else if dataType == common.DataType_FINISH {
				_ = w.finishEventProcess(taskInstance, isSinkOp, inputData)
				logger.Infof("%s finished successfully!", w.SubTaskName)
				break
			} else {
				logger.Errorf("Failed unknown data type")
				return errno.UnknownDataType
			}

			w.pushOutputData(isSinkOp, outputData, dataType, partitionKey)

			//var record *common.Record
			//var dataType common.DataType
			//var dataId string
			//var timeStamp uint64
			//var partitionKey int64
			//// TODO: 初始化成 -1？
			//// TODO: 看一下 Partition key的作用，是否一个算子只能有一个输出，结合 OutputDispenser 看一下
			//partitionKey = -1
			//// input 和 output 的数据类型都是 []byte
			//var inputData []byte
			//var outputData []byte
			//
			//if !isSourceOp {
			//	// 不是 source op
			//	record = <-inputChannel
			//	dataType = record.DataType
			//	// 注意 这里的 data 是 []byte
			//	inputData = record.Data
			//	// logger.Infof("here input data %s", string(inputData))
			//	dataId = record.DataId
			//	timeStamp = record.Timestamp
			//} else {
			//	// 是 source op
			//	select {
			//	case record = <-inputChannel:
			//		inputData = record.Data
			//		dataType = record.DataType
			//	default:
			//		// SourceOp 没有 event，继续处理
			//		dataId = strconv.FormatInt(int64(generator.GetDataIdGeneratorInstance().Next()), 10)
			//		dataType = common.DataType_PICKLE
			//		timeStamp = timestampUtil.GetTimeStamp()
			//		inputData = nil
			//	}
			//}
			//
			//logger.Infof("%v recv: %s(type=%v)", w.SubTaskName, string(inputData), dataType)
			//
			//if dataType == common.DataType_PICKLE {
			//	outputData, partitionKey, _ = w.pickleDataProcess(taskInstance, isKeyOp, inputData)
			//} else if dataType == common.DataType_CHECKPOINT {
			//	w.checkpointEventProcess(taskInstance, isSinkOp, inputData)
			//	logger.Infof("%s success save checkpoint state", w.SubTaskName)
			//	checkpointRecord := &common.Record_Checkpoint{}
			//	checkpointRecord.Unmarshal(inputData)
			//	cancelJob := checkpointRecord.CancelJob
			//	if cancelJob == true {
			//		logger.Infof("%s success finish job", w.SubTaskName)
			//		break
			//	} else {
			//		continue
			//	}
			//} else if dataType == common.DataType_FINISH {
			//	w.finishEventProcess(taskInstance, isSinkOp, inputData)
			//	logger.Infof("%s finished successfully", w.SubTaskName)
			//	break
			//} else {
			//	logger.Fatalf("Failed: unknown data type: %v", dataType)
			//}
			//
			//// TODO 看看 outputchannel 对不对
			//w.pushOutputData(isSinkOp, outputData, dataType, dataId, timeStamp, partitionKey)
			//if !isSinkOp {
			//	// TODO(qiu) 将 output_data 转成 record
			//	output := &common.Record{
			//		DataId:       dataId,
			//		DataType:     dataType,
			//		Data:         outputData,
			//		Timestamp:    timeStamp,
			//		PartitionKey: partitionKey,
			//	}
			//	logger.Infof("output channel %v", output)
			//	outputChannel <- output
			//} else {
			//	logger.Infof("sink word %v", string(inputData))
			//}

			//if isKeyOp {
			//	outputData = inputData
			//	if dataType == common.DataType_PICKLE {
			//		keyBytes, err := taskInstance.Compute(inputData)
			//		if err != nil {
			//			logger.Errorf("Compute error: %v", err)
			//			return err
			//		}
			//		buf := bytes.NewBuffer(keyBytes)
			//		// 小端存储
			//		// TODO：相应的 keyby op 转 bytes 的时候，也要转成小端存储
			//		err = binary.Read(buf, binary.LittleEndian, &partitionKey)
			//		if err != nil {
			//			logger.Errorf("binary.Read error: %v", err)
			//			return err
			//		}
			//	} else if dataType == common.DataType_CHECKPOINT {
			//		// TODO 这里还没处理完
			//		w.PushCheckpointEventToOutputChannel(inputData, outputChannel)
			//		_ = taskInstance.Checkpoint()
			//		logger.Infof("%s success save snapshot state", w.SubTaskName)
			//
			//		tmp := &common.Record_Checkpoint{}
			//		err := tmp.Unmarshal(inputData)
			//		if err != nil {
			//			logger.Errorf("Fail to unmarchal checkpoint data: %v", err)
			//			return err
			//		}
			//
			//		err = w.acknowledgeCheckpoint(w.jobId, tmp.Id, 0, "")
			//		if err != nil {
			//			logger.Errorf("Fail to acknowledge checkpoint: %v", err)
			//			return err
			//		}
			//		if tmp.CancelJob == true {
			//			logger.Infof("%s success finish job", w.SubTaskName)
			//			break
			//		} else {
			//			continue
			//		}
			//	} else if dataType == common.DataType_FINISH {
			//		w.PushFinishEventToOutputChannel(outputChannel)
			//		logger.Infof("%s finished successfully", w.SubTaskName)
			//		break
			//	} else {
			//		logger.Fatalf("Failed: unknown data type: %v", dataType)
			//	}
			//} else {
			//	if dataType == common.DataType_PICKLE {
			//		data, err := taskInstance.Compute(inputData)
			//		if err != nil {
			//			if err == errno.JobFinished {
			//				// TODO(qiu): 通知后续的算子，处理结束
			//				return err
			//			}
			//			logger.Errorf("subtask: %v Compute failed", w.SubTaskName)
			//			return err
			//		}
			//		outputData = data
			//	} else if dataType == common.DataType_CHECKPOINT {
			//		if !isSinkOp {
			//			w.PushCheckpointEventToOutputChannel(inputData, outputChannel)
			//		}
			//		// 把 inputData 转成 record    []byte -> checkpoint
			//		t := &common.Record_Checkpoint{}
			//		err := t.Unmarshal(inputData)
			//		if err != nil {
			//			logger.Errorf("Fail to unmarchal checkpoint data: %v", err)
			//			return err
			//		}
			//		checkpointState := taskInstance.Checkpoint()
			//
			//		err = w.saveCheckpointState(checkpointState, t.Id)
			//		if err != nil {
			//			logger.Errorf("Fail to save checkpoint state: %v", err)
			//			return err
			//		}
			//		logger.Infof("%s success save snapshot state", w.SubTaskName)
			//		// TODO: 验证 t.id 是否是 checkpoint id
			//		err = w.acknowledgeCheckpoint(w.jobId, t.Id, 0, "")
			//		if err != nil {
			//			logger.Errorf("Fail to acknowledge checkpoint: %v", err)
			//			return err
			//		}
			//		if t.CancelJob == true {
			//			logger.Infof("%s cancel job", w.SubTaskName)
			//			break
			//		} else {
			//			continue
			//		}
			//	} else if dataType == common.DataType_FINISH {
			//		logger.Infof("%s finished successfully", w.SubTaskName)
			//		if !isSinkOp {
			//			w.PushFinishEventToOutputChannel(outputChannel)
			//			break
			//		}
			//	} else {
			//		logger.Fatalf("Failed: unknown data type: %v", dataType)
			//	}
			//}
		}
	}
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
	select {
	case record := <-w.InputChannel:
		// 获取到 event，处理event
		// event 可能是 checkpoint 或 finish

		// 只需要返回 DataType, 后面的几个参数没有意义
		return record.DataType, nil
	default:
		// SourceOp 没有 event，继续处理
		// TODO 1.30 dataId 生成的位置要改一下，不在这里生成
		dataType := common.DataType_BINARY
		// data 为 nil，真实的 data 在 执行source算子 compute 的时候填充
		return dataType, nil
	}
}

// -------------------- process data --------------------

// BinaryDataProcess 处理序列化的二进制数据
func (w *Worker) BinaryDataProcess(taskInstance operators.OperatorBase, isKeyOp bool, inputData []byte) ([]byte, int64, error) {
	var outputData []byte = nil
	var partitionKey int64 = -1

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
	} else {
		// 不是 partition key operator, 计算 outputData
		// PS: Source Operator 也会走这里
		outputData, _ = taskInstance.Compute(inputData)

	}
	return outputData, partitionKey, nil
}

//// TODO 结合调用这个函数的位置看一下，看看 inputdata 是不是 Record_Checkpoint 和 finished Record 怎么塞进 Record 里，
//func (w *Worker) checkpointEventProcess(taskInstance operators.OperatorBase, isSinkOp bool, inputData []byte) error {
//	// TODO: 实现无状态算子的 checkpoint 判断
//	if !isSinkOp {
//		// 如果不是 SinkOp，需要把 CheckPoint Event 发送到后继算子
//		w.PushCheckpointEventToOutputChannel(inputData, w.outputChannel)
//	}
//	// 把 inputData 转成 checkpointRecord    []byte -> checkpoint
//	checkpointRecord := &common.Record_Checkpoint{}
//	err := checkpointRecord.Unmarshal(inputData)
//	if err != nil {
//		logger.Errorf("Fail to unmarchal checkpoint data: %v", err)
//		return err
//	}
//	//err = w.checkpoint(taskInstance, checkpointRecord.Id, w.jobId)
//	if err != nil {
//		logger.Errorf("Fail to checkpoint: %v", err)
//		return err
//	}
//	return err
//}

func (w *Worker) finishEventProcess(taskInstance operators.OperatorBase, isSinkOp bool, inputData []byte) error {
	if isSinkOp {
		return nil
	}
	w.PushFinishEventToOutputChannel(w.OutputChannel)
	return nil
}

//func (w *Worker) checkpoint(taskInstance operators.OperatorBase, checkpointId uint64, jobId string) error {
//	// checkpointState 是一个 []byte
//	checkpointState := taskInstance.Checkpoint()
//	// 保存 checkpointState 是一个 []byte
//	err := w.saveCheckpointState(checkpointState, checkpointId)
//	if err != nil {
//		logger.Errorf("Fail to save checkpoint state: %v", err)
//		return err
//	}
//	err = w.acknowledgeCheckpoint(jobId, checkpointId, checkpointState, 0, "")
//	if err != nil {
//		logger.Errorf("Fail to acknowledge checkpoint: %v", err)
//		return err
//	}
//	return nil
//}
//
//func getCheckpointPath(checkpointDir string, checkpointId uint64) string {
//	return path.Join(checkpointDir, fmt.Sprintf("chk_%d", checkpointId))
//}

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

//// 1. checkpoint 的时候，需要将 checkpoint 的数据写入到文件中
//// TODO 2. checkpoint 的时候，存储到卫星的相邻节点上
//func (w *Worker) saveCheckpointState(checkpointState []byte, checkpointId uint64) error {
//	filePath := getCheckpointPath(w.CheckpointDir, checkpointId)
//	return common2.SaveBytesToFile(checkpointState, filePath)
//}

// PushRecord 用于处理 OutputDispatcher 的 Rpc 请求
// 将数据写入 Worker 的 InputReceiver
func (w *Worker) PushRecord(record *common.Record, fromSubTask string, partitionIdx int64) error {
	preSubTask := fromSubTask
	logger.Infof("Recv data(from=%s): %v", preSubTask, record)
	w.InputReceiver.RecvData(partitionIdx, record)
	return nil
}

func (w *Worker) PushEventRecordToOutputChannel(dataId uint64, dataType common.DataType,
	data []byte, outputChannel chan *common.Record) {
	record := &common.Record{
		DataType:     dataType,
		Data:         data,
		PartitionKey: -1,
	}
	outputChannel <- record
}

func (w *Worker) PushFinishEventToOutputChannel(outputChannel chan *common.Record) {
	w.PushEventRecordToOutputChannel(lastFinishDataId, common.DataType_FINISH, nil, outputChannel)
}

//// TODO 在发送第一个 checkpoint event的时候，就应该把checkpointEvent转成bytes
//func (w *Worker) PushCheckpointEventToOutputChannel(checkpointEventBytes []byte,
//	outputChannel chan *common.Record) {
//	w.PushEventRecordToOutputChannel(checkpointDataId, common.DataType_CHECKPOINT,
//		checkpointEventBytes, outputChannel)
//	return
//}

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

//func (w *Worker) TriggerCheckpoint(checkpoint *common.Record_Checkpoint) error {
//	// 只有 SourceOp 才会被调用该函数
//	checkpointData, err := checkpoint.Marshal()
//	if err != nil {
//		logger.Errorf("checkpoint.Marshal error: %v", err)
//		return err
//	}
//	data := &common.Record{
//		DataId:       checkpointDataId,
//		DataType:     common.DataType_CHECKPOINT,
//		Data:         checkpointData,
//		Timestamp:    timestampUtil.GetTimeStamp(),
//		PartitionKey: -1,
//	}
//	w.inputChannel <- data
//	return nil
//}

//// errCode 默认值为0， ErrMsg 默认值为 ""
//func (w *Worker) acknowledgeCheckpoint(jobId string, checkpointId uint64, checkpointState []byte, errCode int32, errMsg string) error {
//	conn, err := messenger.GetRpcConn(w.jobManagerHost, w.jobManagerPort)
//	if err != nil {
//		logger.Errorf("GetRpcConn error: %v", err)
//		return err
//	}
//	defer conn.Close()
//
//	client := sun.NewSunClient(conn)
//	_, err = client.AcknowledgeCheckpoint(context.Background(), &sun.AcknowledgeCheckpointRequest{
//		SubtaskName: w.SubTaskName,
//		JobId:       jobId,
//		//CheckpointId: checkpointId,
//		State: &common.File{
//			Name:    getCheckpointPath(w.CheckpointDir, checkpointId),
//			Content: checkpointState,
//		},
//		Status: &common.Status{
//			ErrCode: errCode,
//			Message: errMsg,
//		},
//	})
//
//	if err != nil {
//		logger.Errorf("AcknowledgeCheckpoint error: %v", err)
//		return err
//	}
//
//	return nil
//}

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
	//worker.CheckpointDir = fmt.Sprintf("tmp/tm/%d/%s/%d/checkpoint", worker.satelliteName, worker.SubTaskName, worker.partitionIdx)
	return worker
}
