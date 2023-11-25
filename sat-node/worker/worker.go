package worker

/*
	slot 中实际执行的 subtask(worker)
*/

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"path"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/operators"
	common2 "satweave/utils/common"
	"satweave/utils/errno"
	"satweave/utils/generator"
	"satweave/utils/logger"
	timestampUtil "satweave/utils/timestamp"
	"sync"
)

type specialDataId int64

const ()

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
	SubTaskName     string
	partitionIdx    int64

	inputReceiver   *InputReceiver
	outputDispenser *OutputDispenser

	cls operators.OperatorBase

	inputChannel  chan *common.Record
	outputChannel chan *common.Record

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.Mutex

	CheckpointDir string

	// job manager endpoint
	jobManagerHost string
	jobManagerPort uint64
	jobId          string

	// 初始化时的 state
	state *common.File
}

func (w *Worker) startComputeOnStandletonProcess() error {
	err := w.ComputeCore(w.inputChannel, w.outputChannel, w.state)
	if err != nil {
		return err
	}
	return nil
}

// ----------------------------------- init for start subtask -----------------------------------
func (w *Worker) initForStartService() {
	if !w.isSourceOp() {
		w.inputChannel = w.initInputReceiver(w.inputEndpoints)
	} else {
		// 如果是 SourceOp
		//这个Channel 用来接收 checkpoint 等 event
		w.inputChannel = make(chan *common.Record, 100)
	}
	if !w.isSinkOp() {
		w.outputChannel = w.initOutputDispenser(w.outputEndpoints)
	}

	// init operator
	w.cls.Init(make(map[string]string))
	w.cls.SetName(w.SubTaskName)
}

func (w *Worker) initInputReceiver(inputEndpoints []*common.InputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	inputChannel := make(chan *common.Record, 10000)
	w.inputReceiver = NewInputReceiver(w.ctx, inputChannel, inputEndpoints)
	return inputChannel
}

func (w *Worker) initOutputDispenser(outputEndpoints []*common.OutputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	outputChannel := make(chan *common.Record, 10000)
	// TODO(qiu): 研究一下 PartitionIdx 的作用
	w.outputDispenser = NewOutputDispenser(w.ctx, outputChannel, outputEndpoints, w.SubTaskName, w.partitionIdx)
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

func (w *Worker) ComputeCore(inputChannel, outputChannel chan *common.Record, state *common.File) error {
	// TODO(qiu):SourceOp 中通过 StopIteration 异常（迭代器终止）来表示
	// 用 error 代替？
	//errChan := make(chan error, 1)
	//go func() {
	//	err := w.innerComputeCore(inputChannel, outputChannel, state*common.File)
	//	if err != nil {
	//		logger.Errorf("innerComputeCore error: %v", err)
	//		errChan <- err
	//		return
	//	}
	//	errChan <- nil
	//}()
	//
	//err := <-errChan
	//if err != nil {
	//	logger.Errorf("innerComputeCore error: %v", err)
	//	if err == errno.JobFinished {
	//		logger.Infof("%s finished successfully", w.SubTaskName)
	//		w.PushFinishEventToOutputChannel(outputChannel)
	//	} else {
	//		logger.Errorf("Failed: run %s task failed, err %v", w.SubTaskName, err)
	//		return err
	//	}
	//}
	err := w.innerComputeCore(inputChannel, outputChannel, state)
	if err == errno.JobFinished {
		// SourceOp 中通过 JobFinishedError 异常来表示
		// 处理完成。向下游 Operator 发送 Finish Record
		logger.Infof("%s finished successfully!", w.SubTaskName)
		w.PushFinishEventToOutputChannel(outputChannel)
	} else {
		logger.Errorf("run %s task failed, err %v", w.SubTaskName, err)
		return err
	}
	return nil
}

func (w *Worker) innerComputeCore(inputChannel, outputChannel chan *common.Record, state *common.File) error {
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
	taskInstance.Init(make(map[string]string))

	// ------------------------ 如果 state 不为空，需要恢复状态 ------------------------
	if state != nil {
		// TODO: taskInstance 需要恢复状态
		// 恢复状态 ack，其实也可不用，状态还没恢复的话，不会处理数据
	}

	for {
		select {
		case <-w.ctx.Done():
			logger.Infof("finish cls %v", w.clsName)
			return nil
		default:
			var partitionKey int64 = -1
			var outputData []byte = nil

			dataType, inputData, dataId, timeStamp := w.getInputData(isSourceOp)

			// TODO 都还没做 error 的处理
			if dataType == common.DataType_PICKLE {
				outputData, partitionKey, _ = w.pickleDataProcess(taskInstance, isKeyOp, inputData)
			} else if dataType == common.DataType_CHECKPOINT {
				_ = w.checkpointEventProcess(taskInstance, isSinkOp, inputData)
				logger.Infof("%s success save checkpoint state", w.SubTaskName)
				// 把 inputdata 转成 checkpoint
				checkpointRecord := &common.Record_Checkpoint{}
				err := checkpointRecord.Unmarshal(inputData)
				if err != nil {
					logger.Errorf("Fail to unmarchal checkpoint data: %v", err)
					return err
				}
				cancelJob := checkpointRecord.CancelJob
				if cancelJob {
					logger.Infof("%s success finish job", w.SubTaskName)
					break
				} else {
					continue
				}
			} else if dataType == common.DataType_FINISH {
				_ = w.finishEventProcess(taskInstance, isSinkOp, inputData)
				logger.Infof("%s finished successfully!", w.SubTaskName)
				break
			} else {
				logger.Errorf("Failed unknown data type")
				return errno.UnknownDataType
			}

			w.pushOutputData(isSinkOp, outputData, dataType, dataId, timeStamp, partitionKey)

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

// -------------------- get input --------------------
func (w *Worker) getInputData(isSourceOp bool) (common.DataType, []byte, int64, uint64) {
	// 不是 SourceOp, 从 inputChannel 正常获取数据
	if !isSourceOp {
		record := <-w.inputChannel
		return record.DataType, record.Data, record.DataId, record.Timestamp
	}

	// 是SourceOp, 从 inputChannel 里获取 event
	select {
	case record := <-w.inputChannel:
		// 获取到 event 处理，event
		// TODO 这里的 -1 -1 不好，需要改
		return record.DataType, record.Data, -1, -1
	default:
		// SourceOp 没有 event，继续处理
		dataId := generator.GetDataIdGeneratorInstance().Next()
		dataType := common.DataType_PICKLE
		timeStamp := timestampUtil.GetTimeStamp()
		// TODO: 为啥是 nil
		// TODO: 如果没有数据，需要往后传播吗？
		// TODO: 这里的data为空，是否需要调用 compute 获取data？
		// TODO: dataId 为 string 是否合理？
		return dataType, nil, dataId, timeStamp
	}
}

// -------------------- process data --------------------
func (w *Worker) pickleDataProcess(taskInstance operators.OperatorBase, isKeyOp bool, inputData []byte) ([]byte, int64, error) {
	var outputData []byte = nil
	var partitionKey int64 = -1

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

// TODO 结合调用这个函数的位置看一下，看看 inputdata 是不是 Record_Checkpoint 和 finished Record 怎么塞进 Record 里，
func (w *Worker) checkpointEventProcess(taskInstance operators.OperatorBase, isSinkOp bool, inputData []byte) error {
	// TODO: 实现无状态算子的 checkpoint 判断
	if !isSinkOp {
		// 如果不是 SinkOp，需要把 CheckPoint Event 推到下一个算子
		// TODO: 看一下 Partition key的作用，是否一个算子只能有一个输出，结合 OutputDispenser 看一下
		w.PushCheckpointEventToOutputChannel(inputData, w.outputChannel)
	}
	// 把 inputData 转成 checkpointRecord    []byte -> checkpoint
	checkpointRecord := &common.Record_Checkpoint{}
	err := checkpointRecord.Unmarshal(inputData)
	if err != nil {
		logger.Errorf("Fail to unmarchal checkpoint data: %v", err)
		return err
	}
	err = w.checkpoint(taskInstance, checkpointRecord.Id, w.jobId)
	if err != nil {
		logger.Errorf("Fail to checkpoint: %v", err)
		return err
	}
	return err
}

func (w *Worker) finishEventProcess(taskInstance operators.OperatorBase, isSinkOp bool, inputData []byte) error {
	if isSinkOp {
		return nil
	}
	w.PushFinishEventToOutputChannel(w.outputChannel)
	return nil
}

func (w *Worker) checkpoint(taskInstance operators.OperatorBase, checkpointId uint64, jobId string) error {
	// checkpointState 是一个 []byte
	checkpointState := taskInstance.Checkpoint()
	// 本地保存 checkpointState 是一个 []byte
	err := w.saveCheckpointState(checkpointState, checkpointId)
	if err != nil {
		logger.Errorf("Fail to save checkpoint state: %v", err)
		return err
	}
	err = w.acknowledgeCheckpoint(jobId, checkpointId, nil, 0, "")
	if err != nil {
		logger.Errorf("Fail to acknowledge checkpoint: %v", err)
		return err
	}
	return nil
}

func getCheckpointPath(checkpointDir string, checkpointId uint64) string {
	return path.Join(checkpointDir, fmt.Sprintf("chk_%d", checkpointId))
}

// -------------------- push output --------------------
func (w *Worker) pushOutputData(isSinkOp bool, outputData []byte, dataType common.DataType, dataId int64, timestamp uint64, partitionKey int64) {
	if isSinkOp {
		return
	}
	record := &common.Record{
		DataId:       dataId,
		DataType:     dataType,
		Data:         outputData,
		Timestamp:    timestamp,
		PartitionKey: partitionKey,
	}
	// TODO 看看这个 output 有没写错
	// TODO data 需要再封装一层，结合 output 的逻辑看一下
	w.outputChannel <- record
}

// 1. checkpoint 的时候，需要将 checkpoint 的数据写入到文件中
// TODO 2. checkpoint 的时候，存储到卫星的相邻节点上
func (w *Worker) saveCheckpointState(checkpointState []byte, checkpointId uint64) error {
	filePath := getCheckpointPath(w.CheckpointDir, checkpointId)
	return common2.SaveBytesToFile(checkpointState, filePath)
}

func (w *Worker) PushRecord(record *common.Record, fromSubTask string, partitionIdx int64) error {
	preSubTask := fromSubTask
	logger.Infof("Recv data(from=%s): %v", preSubTask, record)
	w.inputReceiver.RecvData(partitionIdx, record)
	return nil
}

func (w *Worker) PushEventRecordToOutputChannel(dataId string, dataType common.DataType,
	data []byte, outputChannel chan *common.Record) {
	record := &common.Record{
		DataId:       dataId,
		DataType:     dataType,
		Data:         data,
		Timestamp:    timestampUtil.GetTimeStamp(),
		PartitionKey: -1,
	}
	outputChannel <- record
}

func (w *Worker) PushFinishEventToOutputChannel(outputChannel chan *common.Record) {
	record := &common.Record{
		DataId:    "last_finish_data_id",
		DataType:  common.DataType_FINISH,
		Timestamp: timestampUtil.GetTimeStamp(),
		// TODO(qiu): 这里Data是空的，需要修改
		Data:         nil,
		PartitionKey: -1,
	}
	outputChannel <- record
}

func (w *Worker) PushCheckpointEventToOutputChannel(checkpoint []byte,
	outputChannel chan *common.Record) {
	w.PushEventRecordToOutputChannel("checkpoint_data_id", common.DataType_CHECKPOINT,
		checkpoint, outputChannel)
	return
}

// Run 启动 Worker
func (w *Worker) Run() {
	// TODO(qiu)：都还需要添加错误处理
	// 启动 Receiver 和 Dispenser
	if w.inputReceiver != nil {
		w.inputReceiver.RunAllPartitionReceiver()
		logger.Infof("subtask %v start input receiver finish...", w.SubTaskName)
	}

	if w.outputDispenser != nil {
		w.outputDispenser.Run()
		logger.Infof("subtask %v start output dispenser finish...", w.SubTaskName)
	}

	// 启动 worker 核心进程
	w.startComputeOnStandletonProcess()
	//time.Sleep(60 * time.Second)
	logger.Infof("start core compute process success...")
}

func (w *Worker) Stop() {
	w.cancel()
}

func (w *Worker) TriggerCheckpoint(checkpoint *common.Record_Checkpoint) error {
	// 只有 SourceOp 才会被调用该函数
	checkpointData, err := checkpoint.Marshal()
	if err != nil {
		logger.Errorf("checkpoint.Marshal error: %v", err)
		return err
	}
	data := &common.Record{
		DataId:       "checkpoint_data_id",
		DataType:     common.DataType_CHECKPOINT,
		Data:         checkpointData,
		Timestamp:    timestampUtil.GetTimeStamp(),
		PartitionKey: -1,
	}
	w.inputChannel <- data
	return nil
}

// errCode 默认值为0， ErrMsg 默认值为 ""
func (w *Worker) acknowledgeCheckpoint(jobId string, checkpointId uint64, checkpointState []byte, errCode int32, errMsg string) error {
	conn, err := messenger.GetRpcConn(w.jobManagerHost, w.jobManagerPort)
	if err != nil {
		logger.Errorf("GetRpcConn error: %v", err)
		return err
	}
	defer conn.Close()

	client := sun.NewSunClient(conn)
	_, err = client.AcknowledgeCheckpoint(context.Background(), &sun.AcknowledgeCheckpointRequest{
		SubtaskName:  w.SubTaskName,
		JobId:        jobId,
		CheckpointId: checkpointId,
		State: &common.File{
			Name:    getCheckpointPath(w.CheckpointDir, checkpointId),
			Content: checkpointState,
		},
		Status: &common.Status{
			ErrCode: errCode,
			Message: errMsg,
		},
	})

	if err != nil {
		logger.Errorf("AcknowledgeCheckpoint error: %v", err)
		return err
	}

	return nil
}

func NewWorker(raftId uint64, executeTask *common.ExecuteTask, jobManagerHost string, jobManagerPort uint64, jobId string) *Worker {
	// TODO(qiu): 这个 ctx 是否可以继承 task manager
	workerCtx, cancel := context.WithCancel(context.Background())
	worker := &Worker{
		raftId:          raftId,
		subtaskId:       uuid.New().String(),
		clsName:         executeTask.ClsName,
		inputEndpoints:  executeTask.InputEndpoints,
		outputEndpoints: executeTask.OutputEndpoints,
		SubTaskName:     executeTask.SubtaskName,
		partitionIdx:    executeTask.PartitionIdx,
		workerId:        executeTask.WorkerId,

		inputReceiver:   nil,
		outputDispenser: nil,

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
	worker.initForStartService()
	worker.CheckpointDir = fmt.Sprintf("tmp/tm/%d/%s/%d/checkpoint", worker.raftId, worker.SubTaskName, worker.partitionIdx)
	return worker
}
