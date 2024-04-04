package operators

import (
	"math/cmplx"
	"satweave/utils/common"
)

// LaSumOp 是低通聚合算子
type LaSumOp struct {
	name    string
	counter uint64
}

func (op *LaSumOp) SetName(name string) {
	op.name = name
}

func (op *LaSumOp) Init(initMap map[string]interface{}) {
	op.counter = initMap["counter"].(uint64)
}

func (op *LaSumOp) Compute(data []byte) ([]byte, error) {
	signalFFT, err := common.BytesToComplex128Slice(data)
	if err != nil {
		return nil, err
	}

	// 低通聚合
	for _, val := range signalFFT {
		if cmplx.Abs(val) > 0 {
			op.counter++
		}
	}

	res := common.Uint64ToBytes(op.counter)

	return res, nil
}

func (op *LaSumOp) IsSourceOp() bool {
	return false
}

func (op *LaSumOp) IsSinkOp() bool {
	return false
}

func (op *LaSumOp) IsKeyByOp() bool {
	return false
}

func (op *LaSumOp) Checkpoint() []byte {
	// TODO: checkpoint
	return nil
}

func (op *LaSumOp) RestoreFromCheckpoint(SunIp, ClsName string, SunPort uint64) error {
	return nil
}
