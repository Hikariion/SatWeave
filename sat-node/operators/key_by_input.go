package operators

type KeyByInputOp struct {
	name string
}

func (op *KeyByInputOp) Init(map[string]string) {

}

func (op *KeyByInputOp) Compute([]byte) (string, error) {
	return "", nil
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
