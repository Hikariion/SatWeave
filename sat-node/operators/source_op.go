package operators

type SourceOperator struct {
	OperatorBase
}

func (op *SourceOperator) Init(map[string]string) {

}

func (op *SourceOperator) Compute([]byte) ([]byte, error) {
	return nil, nil
}

func (op *SourceOperator) IsSourceOp() bool {
	return true
}

func (op *SourceOperator) IsSinkOp() bool {
	return false
}

func (op *SourceOperator) IsKeyByOp() bool {
	return false
}

func (op *SourceOperator) Checkpoint() []byte {
	return nil
}
