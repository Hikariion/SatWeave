package operators

type KeyByInputOp struct {
}

func (op *KeyByInputOp) Compute(data []byte) (string, error) {
	return "", nil
}

func (op *KeyByInputOp) SetName(name string) {

}
