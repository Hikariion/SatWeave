package operators

type KeyByInputOp struct {
	JobId string
}

func (op *KeyByInputOp) Init(initMap map[string]interface{}) {

}

func (op *KeyByInputOp) Compute([]byte) ([]byte, error) {
	return nil, nil
}

func (op *KeyByInputOp) SetJobId(JobId string) {
	op.JobId = JobId
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

func (op *KeyByInputOp) RestoreFromCheckpoint(SunIp, ClsName string, SunPort uint64) error {
	return nil
}
