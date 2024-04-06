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

type SimpleSink struct {
	JobId        string
	file         *os.File
	nextRecordId uint64
}

func (op *SimpleSink) Init(initMap map[string]interface{}) {
	// 创建一个文件用来存储
	logger.Infof("Init Simple Sink...")
	// 创建一个文件用来存储
	var err error
	op.file, err = os.Create("output.txt")
	if err != nil {
		logger.Fatalf("Failed to create file: %v", err)
	}
}

func (op *SimpleSink) Compute(data []byte) ([]byte, error) {
	str := string(data)
	logger.Infof("%v: %v", op.JobId, str)

	// 把 dataId 转成 str
	dataIdStr := strconv.FormatUint(op.nextRecordId, 10)

	// 将字符串写入文件，每个条目占一行
	if _, err := op.file.WriteString(dataIdStr + " " + str + "\n"); err != nil {
		return nil, err
	}

	op.nextRecordId++
	return nil, nil
}

func (op *SimpleSink) SetJobId(JobId string) {
	op.JobId = JobId
}

func (op *SimpleSink) IsSourceOp() bool {
	return false
}

func (op *SimpleSink) IsSinkOp() bool {
	return true
}

func (op *SimpleSink) IsKeyByOp() bool {
	return false
}

func (op *SimpleSink) Checkpoint() []byte {
	res := common.Uint64ToBytes(op.nextRecordId)
	return res
}

func (op *SimpleSink) RestoreFromCheckpoint(SunIp, ClsName string, SunPort uint64) error {
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
