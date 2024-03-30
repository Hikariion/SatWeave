package operators

import (
	"github.com/mjibson/go-dsp/fft"
	"satweave/utils/common"
)

// FFTOp 是快速傅里叶变换算子
type FFTOp struct {
	OperatorBase
}

func (op *FFTOp) Init(initMap map[string]interface{}) {

}

func (op *FFTOp) Compute(data []byte) ([]byte, error) {
	signal, err := common.BytesToComplex128Slice(data)
	if err != nil {
		return nil, err
	}

	signalFFT := fft.FFT(signal)

	res, err := common.Complex128SliceToBytes(signalFFT)
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

func (op *FFTOp) RestoreFromCheckpoint([]byte) error {
	return nil
}
