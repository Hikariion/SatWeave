package operators

import (
	"github.com/mjibson/go-dsp/fft"
	"satweave/utils/common"
)

// FFTOp 是快速傅里叶变换算子
type FFTOp struct {
	name string
}

func (op *FFTOp) SetName(name string) {
	op.name = name
}

func (op *FFTOp) Init(initMap map[string]interface{}) {

}

func (op *FFTOp) Compute(data []byte) ([]byte, error) {
	signal, err := common.BytesToComplex128Slice(data)
	if err != nil {
		return nil, err
	}

	fft.FFT(signal)

	res, err := common.Complex128SliceToBytes(signal)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (op *FFTOp) IsSourceOp() bool {
	return false
}

func (op *FFTOp) IsSinkOp() bool {
	return false
}

func (op *FFTOp) IsKeyByOp() bool {
	return false
}

func (op *FFTOp) Checkpoint() []byte {
	// TODO: checkpoint
	return nil
}

func (op *FFTOp) RestoreFromCheckpoint(SunIp, ClsName string, SunPort uint64) error {
	return nil
}
