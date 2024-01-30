package operators

type SumOp struct {
	OperatorBase
	counter map[string]uint64
}

func (op *SumOp) Init() {
	op.counter = make(map[string]uint64)
}

func (op *SumOp) Compute(data []byte) ([]byte, error) {
	dataStr := string(data)
	if _, ok := op.counter[dataStr]; !ok {
		op.counter[dataStr] = 0
	}
	op.counter[dataStr]++
	return nil, nil
}

func (op *SumOp) IsSourceOp() bool {
	return false
}

func (op *SumOp) IsSinkOp() bool {
	return false
}

func (op *SumOp) IsKeyByOp() bool {
	return false
}

func (op *SumOp) Checkpoint() []byte {
	// TODO: checkpoint
	return nil
}

func (op *SumOp) RestoreFromCheckpoint([]byte) error {
	return nil
}
