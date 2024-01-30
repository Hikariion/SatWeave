package operators

type KeyByInputOp struct {
	name string
}

func (op *KeyByInputOp) Init([]byte) {

}

func (op *KeyByInputOp) Compute([]byte) ([]byte, error) {
	return nil, nil
}

func (op *KeyByInputOp) SetName(name string) {
	op.name = name
}

func (op *KeyByInputOp) IsSourceOp() bool {
	return false
}

func (op *KeyByInputOp) IsSinkOp() bool {
	return false
}

func (op *KeyByInputOp) IsKeyByOp() bool {
	return true
}

func (op *KeyByInputOp) Checkpoint() []byte {
	return nil
}

func (op *KeyByInputOp) RestoreFromCheckpoint([]byte) error {
	return nil
}
