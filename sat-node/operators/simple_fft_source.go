package operators

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"satweave/cloud/sun"
	"satweave/messenger"
	common2 "satweave/messenger/common"
	"satweave/utils/common"
	"satweave/utils/logger"
	"strconv"
	"strings"
	"time"
)

const offset = 100

type SimpleFFTSource struct {
	JobId string
	// 用于 Source 算子和 Worker 之间通信
	InputChannel chan *common2.Record
	nextRecordId uint64

	// For dsp
	fs    float64
	T     float64
	fLow  float64
	fHigh float64
}

func (op *SimpleFFTSource) Init(initMap map[string]interface{}) {
	op.InputChannel = initMap["InputChannel"].(chan *common2.Record)

	op.fs = 1000.0
	op.T = 1.0
	op.fHigh = 50.0
	op.fLow = 5.0

	endId := op.nextRecordId + offset
	checkpointId := op.nextRecordId + uint64(offset*0.8)

	go func() {
		for {
			signal, err := readSignalByID("test-files/signals.txt", op.nextRecordId)

			if err != nil {
				record := &common2.Record{
					DataType: common2.DataType_FINISH,
					Data:     nil,
				}

				op.InputChannel <- record

				return
			}

			data, err := common.Complex128SliceToBytes(signal)
			if err != nil {
				return
			}

			record := &common2.Record{
				DataType: common2.DataType_BINARY,
				Data:     data,
			}

			op.InputChannel <- record

			op.nextRecordId++

			if op.nextRecordId == endId {
				return
			}
			if op.nextRecordId == checkpointId {
				record := &common2.Record{
					DataType: common2.DataType_CHECKPOINT,
					Data:     nil,
				}
				op.InputChannel <- record
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func (op *SimpleFFTSource) Compute([]byte) ([]byte, error) {
	return nil, nil
}

func (op *SimpleFFTSource) SetJobId(JobId string) {
	op.JobId = JobId
}

func (op *SimpleFFTSource) IsSourceOp() bool {
	return true
}

func (op *SimpleFFTSource) IsSinkOp() bool {
	return false
}

func (op *SimpleFFTSource) IsKeyByOp() bool {
	return false
}

func (op *SimpleFFTSource) Checkpoint() []byte {
	res := common.Uint64ToBytes(op.nextRecordId)
	return res
}

func (op *SimpleFFTSource) RestoreFromCheckpoint(SunIp, ClsName string, SunPort uint64) error {
	conn, err := messenger.GetRpcConn(SunIp, SunPort)
	if err != nil {
		logger.Errorf("Fail to get rpc conn on TaskManager %v", SunIp)
		return err
	}
	client := sun.NewSunClient(conn)
	result, err := client.RestoreFromCheckpoint(context.Background(),
		&sun.RestoreFromCheckpointRequest{
			SubtaskName: ClsName,
		})
	if err != nil {
		return err
	}
	state := result.State
	if state == nil {
		op.nextRecordId = 0
	} else {
		op.nextRecordId = common.BytesToUint64(state)
	}
	return nil
}

func readSignalByID(filePath string, id uint64) ([]complex128, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var foundID bool
	var signal []complex128

	for scanner.Scan() {
		line := scanner.Text()

		// 尝试将行转换为ID
		currentID, err := strconv.Atoi(line)
		if err == nil {
			if currentID == int(id) {
				foundID = true // 找到匹配的ID
				continue       // 继续读取下一行以获取信号数据
			}
		}

		if foundID {
			// 找到ID后，当前行包含信号数据
			dataParts := strings.Split(line, ", ")
			for _, part := range dataParts {
				values := strings.Split(part, " ")
				if len(values) != 2 {
					continue // 如果数据格式不符，跳过此数据点
				}
				realPart, err := strconv.ParseFloat(values[0], 64)
				if err != nil {
					continue // 如果实部无法转换，跳过此数据点
				}
				imagPart, err := strconv.ParseFloat(values[1], 64)
				if err != nil {
					continue // 如果虚部无法转换，跳过此数据点
				}
				signal = append(signal, complex(realPart, imagPart))
			}
			return signal, nil // 返回找到的信号
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if !foundID {
		return nil, fmt.Errorf("signal not found for ID %d", id)
	}
	return signal, nil
}
