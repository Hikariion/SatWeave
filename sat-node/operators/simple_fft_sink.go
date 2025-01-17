package operators

import (
	"context"
	"os"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/utils/common"
	"satweave/utils/logger"
	"strconv"
)

type SimpleFFTSink struct {
	name          string
	file          *os.File
	nextRecordId  uint64
	jobId         string
	SunIp         string
	SunPort       uint64
	SatelliteName string
}

func (op *SimpleFFTSink) Init(initMap map[string]interface{}) {
	op.SunIp = initMap["SunIp"].(string)
	op.SunPort = initMap["SunPort"].(uint64)
	op.SatelliteName = initMap["SatelliteName"].(string)
	// 创建一个文件用来存储
	logger.Infof("Init Simple FFT Sink...")
	// 创建一个文件用来存储
	var err error
	op.file, err = os.Create("fft_output.txt")
	if err != nil {
		logger.Fatalf("Failed to create file: %v", err)
	}
}

func (op *SimpleFFTSink) Compute(data []byte) ([]byte, error) {
	res := common.BytesToUint64(data)
	logger.Infof("SimpleFFTSink  %d", res)

	// 把 res 转成 string
	resStr := strconv.FormatUint(res, 10)
	resStr = resStr + " " + op.SatelliteName

	// 把 dataId 转成 str
	dataIdStr := strconv.FormatUint(op.nextRecordId, 10)

	conn, err := messenger.GetRpcConn(op.SunIp, op.SunPort)
	if err != nil {
		logger.Errorf("Fail to get rpc conn on TaskManager %v", op.SunIp)
		return nil, err
	}
	client := sun.NewSunClient(conn)
	_, err = client.ReceiverStreamData(context.Background(),
		&sun.ReceiverStreamDataRequest{
			JobId:  op.jobId,
			DataId: dataIdStr,
			Res:    resStr,
		})

	//// 将 res 写入文件，每个条目占一行
	//if _, err := op.file.WriteString(dataIdStr + " " + resStr + "\n"); err != nil {
	//	return nil, err
	//}

	op.nextRecordId++

	return nil, nil
}

func (op *SimpleFFTSink) SetJobId(jobId string) {
	op.jobId = jobId
}

func (op *SimpleFFTSink) IsSourceOp() bool {
	return false
}

func (op *SimpleFFTSink) IsSinkOp() bool {
	return true
}

func (op *SimpleFFTSink) IsKeyByOp() bool {
	return false
}

func (op *SimpleFFTSink) Checkpoint() []byte {
	res := common.Uint64ToBytes(op.nextRecordId)
	return res
}

func (op *SimpleFFTSink) RestoreFromCheckpoint(SunIp, SubTaskName string, SunPort uint64) error {
	conn, err := messenger.GetRpcConn(SunIp, SunPort)
	if err != nil {
		logger.Errorf("Fail to get rpc conn on TaskManager %v", SunIp)
		return err
	}
	client := sun.NewSunClient(conn)
	result, err := client.RestoreFromCheckpoint(context.Background(),
		&sun.RestoreFromCheckpointRequest{
			SubtaskName: SubTaskName,
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
