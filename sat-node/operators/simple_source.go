package operators

import (
	"bufio"
	"os"
	"satweave/messenger/common"
	"satweave/utils/logger"
)

type SimpleSource struct {
	name string

	// 用于 Source 算子和 Worker 之间的通信
	InputChannel chan *common.Record
}

func (op *SimpleSource) Init(initMap map[string]interface{}) {
	op.InputChannel = initMap["InputChannel"].(chan *common.Record)

	go func() {
		file, err := os.Open("./test-files/document.txt")
		if err != nil {
			logger.Errorf("open file failed: %v", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			word := scanner.Text()
			if word != "" && len(word) > 0 {
				// TODO 1.30 应该在这个地方插入 Event，而不是在 Compute 中
				record := &common.Record{
					DataType: common.DataType_BINARY,
					Data:     []byte(word),
				}
				op.InputChannel <- record
			}
		}
		return
	}()
}

func (op *SimpleSource) Compute([]byte) ([]byte, error) {
	return nil, nil
}

func (op *SimpleSource) SetName(name string) {
	op.name = name
}

func (op *SimpleSource) IsSourceOp() bool {
	return true
}

func (op *SimpleSource) IsSinkOp() bool {
	return false
}

func (op *SimpleSource) IsKeyByOp() bool {
	return false
}

func (op *SimpleSource) Checkpoint() []byte {
	return nil
}

func (op *SimpleSource) RestoreFromCheckpoint([]byte) error {
	return nil
}
