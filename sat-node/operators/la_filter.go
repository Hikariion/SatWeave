package operators

import (
	"math"
	"satweave/utils/common"
)

// LaFilterOp 是低通滤波算子
type LaFilterOp struct {
	fHigh  float64 // 高频分量
	T      float64 // 信号时长
	cutoff float64 // 截止频率
	fs     float64 // 采样频率
}

func (op *LaFilterOp) Init(initMap map[string]interface{}) {
	op.fHigh = 50.0
	op.T = 1.0
	op.cutoff = 30.0
	op.fs = 1000.0
}

func (op *LaFilterOp) Compute(data []byte) ([]byte, error) {
	signalFFT, err := common.BytesToComplex128Slice(data)
	if err != nil {
		return nil, err
	}

	// 低通滤波：仅保留低于截止频率的频率分量
	for i, _ := range signalFFT {
		freq := float64(i) / op.T
		if freq > op.fs/2 {
			freq = freq - op.fs
		}
		if math.Abs(freq) > op.cutoff {
			signalFFT[i] = 0
		}
	}

	res, err := common.Complex128SliceToBytes(signalFFT)

	if err != nil {
		return nil, err
	}

	return res, nil
}
