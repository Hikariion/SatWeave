package operators

import (
	"math/cmplx"
	"satweave/utils/common"
)

// HaSumOp 是高通聚合算子
type HaSumOp struct {
	OperatorBase
	counter uint64
}

func (op *HaSumOp) Init(initMap map[string]interface{}) {
	op.counter = initMap["counter"].(uint64)
}

func (op *HaSumOp) Compute(data []byte) ([]byte, error) {
	signalFFT, err := common.BytesToComplex128Slice(data)
	if err != nil {
		return nil, err
	}

	// 高通聚合
	for _, val := range signalFFT {
		if cmplx.Abs(val) > 0 {
			op.counter++
		}
	}

	res := common.Uint64ToBytes(op.counter)

	return res, nil
}
