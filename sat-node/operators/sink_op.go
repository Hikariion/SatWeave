package operators

type SinkOperator struct {
	name string
}

func (op *SinkOperator) Init(map[string]string) {

}

func (op *SinkOperator) Compute([]byte) ([]byte, error) {
	return nil, nil
}

func (op *SinkOperator) SetName(name string) {
	op.name = name
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
