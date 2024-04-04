package operators

import (
	"os"
	"satweave/utils/common"
	"satweave/utils/logger"
	"strconv"
)

type SimpleFFTSink struct {
	name string
	file *os.File
}

func (op *SimpleFFTSink) Init(initMap map[string]interface{}) {
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

	// 将 res 写入文件，每个条目占一行
	if _, err := op.file.WriteString(resStr + "\n"); err != nil {
		return nil, err
	}
	return nil, nil
}

func (op *SimpleFFTSink) SetName(name string) {
	op.name = name
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
	return nil
}

func (op *SimpleFFTSink) RestoreFromCheckpoint(SunIp, ClsName string, SunPort uint64) error {
	return nil
}
