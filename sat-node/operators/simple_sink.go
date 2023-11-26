package operators

import "satweave/utils/logger"

type SimpleSink struct {
	name string
}

func (op *SimpleSink) Init(map[string]string) {
	logger.Infof("Init Simple Sink...")
}

func (op *SimpleSink) Compute(data []byte) ([]byte, error) {
	str := string(data)
	logger.Infof("%v: %v", op.name, str)
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
