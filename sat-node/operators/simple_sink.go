package operators

import (
	"os"
	"satweave/utils/logger"
)

type SimpleSink struct {
	name string
	file *os.File
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
	logger.Infof("%v: %v", op.name, str)

	// 将字符串写入文件，每个条目占一行
	if _, err := op.file.WriteString(str + "\n"); err != nil {
		return nil, err
	}
	return nil, nil
}

func (op *SimpleSink) SetName(name string) {
	op.name = name
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
	return nil
}

func (op *SimpleSink) RestoreFromCheckpoint([]byte) error {
	return nil
}
