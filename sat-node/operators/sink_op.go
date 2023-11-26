package operators

type SinkOperator struct {
	OperatorBase
}

func (op *SinkOperator) Init(map[string]string) {

}

func (op *SinkOperator) Compute([]byte) ([]byte, error) {
	return nil, nil
}

func (op *SinkOperator) IsSourceOp() bool {
	return false
}

func (op *SinkOperator) IsSinkOp() bool {
	return true
}

func (op *SinkOperator) IsKeyByOp() bool {
	return false
}

func (op *SinkOperator) Checkpoint() []byte {
	return nil
}

func (op *SinkOperator) RestoreFromCheckpoint([]byte) error {
	return nil
}
