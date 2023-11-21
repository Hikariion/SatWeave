package worker

/*
	slot 中实际执行的 subtask(worker)
*/

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/messenger/common"
	"satweave/sat-node/operators"
	"satweave/utils/errno"
	"satweave/utils/generator"
	"satweave/utils/logger"
	timestampUtil "satweave/utils/timestamp"
	"strconv"
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
}

func (w *Worker) startComputeOnStandletonProcess() error {
	err := w.ComputeCore(w.clsName, w.inputChannel, w.outputChannel)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) initForStartService() {
	if !w.isSourceOp() {
		w.inputChannel = w.initInputReceiver(w.inputEndpoints)
	} else {
		// 用来接收 checkpoint 等 event
		w.inputChannel = make(chan *common.Record, 100)
	}
	if !w.isSinkOp() {
		w.outputChannel = w.initOutputDispenser(w.outputEndpoints)
	}

	// init operator
	w.cls.Init(make(map[string]string))
	w.cls.SetName(w.SubTaskName)
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

func (w *Worker) initInputReceiver(inputEndpoints []*common.InputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	inputChannel := make(chan *common.Record, 10000)
	w.inputReceiver = NewInputReceiver(w.ctx, inputChannel, inputEndpoints)
	return inputChannel
}

func (w *Worker) ComputeCore(clsName string, inputChannel, outputChannel chan *common.Record) error {
	// TODO(qiu):SourceOp 中通过 StopIteration 异常（迭代器终止）来表示
	// 用 error 代替？
	errChan := make(chan error, 1)
	go func() {
		err := w.innerComputeCore(inputChannel, outputChannel)
		if err != nil {
			logger.Errorf("innerComputeCore error: %v", err)
			errChan <- err
			return
		}
		errChan <- nil
	}()

	err := <-errChan
	if err != nil {
		logger.Errorf("innerComputeCore error: %v", err)
		if err == errno.JobFinished {
			logger.Infof("%s finished successfully", w.SubTaskName)
			w.PushFinishEventToOutputChannel(outputChannel)
		} else {
			logger.Errorf("Failed: run %s task failed, err %v", w.SubTaskName, err)
			return err
		}
	}
	return nil
}

func (w *Worker) innerComputeCore(inputChannel, outputChannel chan *common.Record) error {
	logger.Infof("start innerComputeCore...")
	// 具体执行逻辑
	// 判断是否为 source 算子
	isSourceOp := w.isSourceOp()
	// 判断是否为 sink 算子
	isSinkOp := w.isSinkOp()
	// 判断是否为 key 算子
	isKeyOp := w.isKeyOp()

	taskInstance := w.cls
	// TODO(qiu): succ start event?

	for {
		select {
		case <-w.ctx.Done():
			logger.Infof("finish cls %v", w.clsName)
			return nil
		default:
			var record *common.Record
			var dataType common.DataType
			var dataId string
			var timeStamp uint64
			var partitionKey int64
			// TODO: 初始化成 -1？
			partitionKey = -1
			// input 和 output 的数据类型都是 []byte
			var inputData []byte
			var outputData []byte
			if !isSourceOp {
				// 不是 source op
				record = <-inputChannel
				dataType = record.DataType
				// 注意 这里的 data 是 []byte
				inputData = record.Data
				logger.Infof("here input data %s", string(inputData))
				dataId = record.DataId
				timeStamp = record.Timestamp
			} else {
				// 是 source op
				select {
				case record = <-inputChannel:
					inputData = record.Data
					dataType = record.DataType
				default:
					// SourceOp 没有 event，继续处理
					dataId = strconv.FormatInt(int64(generator.GetDataIdGeneratorInstance().Next()), 10)
					dataType = common.DataType_PICKLE
					timeStamp = timestampUtil.GetTimeStamp()
					inputData = nil
				}
			}

			logger.Infof("%v recv: %s(type=%v)", w.SubTaskName, string(inputData), dataType)

			if isKeyOp {
				outputData = inputData
				if dataType == common.DataType_PICKLE {
					keyBytes, err := taskInstance.Compute(inputData)
					if err != nil {
						logger.Errorf("Compute error: %v", err)
						return err
					}
					buf := bytes.NewBuffer(keyBytes)
					// 小端存储
					// TODO：相应的 keyby op 转 bytes 的时候，也要转成小端存储
					err = binary.Read(buf, binary.LittleEndian, &partitionKey)
					if err != nil {
						logger.Errorf("binary.Read error: %v", err)
						return err
					}
				} else if dataType == common.DataType_CHECKPOINT {
					// TODO 这里还没处理完
					w.PushCheckpointEventToOutputChannel(inputData, outputChannel)
					_ = taskInstance.Checkpoint()
					logger.Infof("%s success save snapshot state", w.SubTaskName)

					tmp := &common.Record_Checkpoint{}
					err := tmp.Unmarshal(inputData)
					if err != nil {
						logger.Errorf("Fail to unmarchal checkpoint data: %v", err)
						return err
					}

					err = w.acknowledgeCheckpoint(w.jobId, tmp.Id, 0, "")
					if err != nil {
						logger.Errorf("Fail to acknowledge checkpoint: %v", err)
						return err
					}
					if tmp.CancelJob == true {
						logger.Infof("%s success finish job", w.SubTaskName)
						break
					} else {
						continue
					}
				} else if dataType == common.DataType_FINISH {
					w.PushFinishEventToOutputChannel(outputChannel)
					logger.Infof("%s finished successfully", w.SubTaskName)
					break
				} else {
					logger.Fatalf("Failed: unknown data type: %v", dataType)
				}
			} else {
				if dataType == common.DataType_PICKLE {
					data, err := taskInstance.Compute(inputData)
					if err != nil {
						if err == errno.JobFinished {
							// TODO(qiu): 通知后续的算子，处理结束
							return err
						}
						logger.Errorf("subtask: %v Compute failed", w.SubTaskName)
						return err
					}
					outputData = data
				} else if dataType == common.DataType_CHECKPOINT {
					if !isSinkOp {
						w.PushCheckpointEventToOutputChannel(inputData, outputChannel)
					}
					// 把 inputData 转成 record    []byte -> checkpoint
					t := &common.Record_Checkpoint{}
					err := t.Unmarshal(inputData)
					if err != nil {
						logger.Errorf("Fail to unmarchal checkpoint data: %v", err)
						return err
					}
					checkpointState := taskInstance.Checkpoint()

					err = w.saveCheckpointState(checkpointState, t.Id)
					if err != nil {
						logger.Errorf("Fail to save checkpoint state: %v", err)
						return err
					}
					logger.Infof("%s success save snapshot state", w.SubTaskName)
					// TODO: 验证 t.id 是否是 checkpoint id
					err = w.acknowledgeCheckpoint(w.jobId, t.Id, 0, "")
					if err != nil {
						logger.Errorf("Fail to acknowledge checkpoint: %v", err)
						return err
					}
					if t.CancelJob == true {
						logger.Infof("%s cancel job", w.SubTaskName)
						break
					} else {
						continue
					}
				} else if dataType == common.DataType_FINISH {
					logger.Infof("%s finished successfully", w.SubTaskName)
					if !isSinkOp {
						w.PushFinishEventToOutputChannel(outputChannel)
						break
					}
				} else {
					logger.Fatalf("Failed: unknown data type: %v", dataType)
				}
			}

			if !isSinkOp {
				// TODO(qiu) 将 output_data 转成 record
				output := &common.Record{
					DataId:       dataId,
					DataType:     dataType,
					Data:         outputData,
					Timestamp:    timeStamp,
					PartitionKey: partitionKey,
				}
				logger.Infof("output channel %v", output)
				outputChannel <- output
			} else {
				logger.Infof("sink word %v", string(inputData))
			}
		}
	}
}

func (w *Worker) initOutputDispenser(outputEndpoints []*common.OutputEndpoints) chan *common.Record {
	// TODO(qiu): 调整 channel 容量
	outputChannel := make(chan *common.Record, 10000)
	// TODO(qiu): 研究一下 PartitionIdx 的作用
	w.outputDispenser = NewOutputDispenser(w.ctx, outputChannel, outputEndpoints, w.SubTaskName, w.partitionIdx)
	return outputChannel

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

// 1. checkpoint 的时候，需要将 checkpoint 的数据写入到文件中
// TODO 2. checkpoint 的时候，存储到卫星的相邻节点上
func (w *Worker) saveCheckpointState(checkpointState []byte, checkpointId uint64) error {
	filePath := path.Join(w.CheckpointDir, fmt.Sprintf("chk_%d", checkpointId))

	err := os.WriteFile(filePath, checkpointState, 0644)
	if err != nil {
		logger.Errorf("os.WriteFile error: %v", err)
		return err
	}

	return nil
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
func (w *Worker) acknowledgeCheckpoint(jobId string, checkpointId uint64, errCode int32, errMsg string) error {
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
