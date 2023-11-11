package operators

type SimpleSource struct {
	name string
}

func (op *SimpleSource) Init(map[string]string) {

}

func (op *SimpleSource) Compute([]byte) (string, error) {
	return "", nil
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
