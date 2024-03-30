package operators

import (
	"math"
	common2 "satweave/messenger/common"
	"satweave/utils/common"
	"time"
)

type SimpleFFTSource struct {
	name string
	// 用于 Source 算子和 Worker 之间通信
	InputChannel chan common2.Record

	// For dsp
	fs    float64
	T     float64
	fLow  float64
	fHigh float64
}

func (op *SimpleFFTSource) Init(initMap map[string]interface{}) {
	op.InputChannel = initMap["InputChannel"].(chan common2.Record)

	op.fs = 1000.0
	op.T = 1.0
	op.fHigh = 50.0
	op.fLow = 5.0

	go func() {
		count := 0
		for {
			if count > 20 {
				return
			}
			N := int(op.fs * op.T)
			signal := make([]complex128, N)
			for n := 0; n < N; n++ {
				t := float64(n) / op.fs
				signal[n] = complex(math.Sin(2*math.Pi*op.fLow*t)+0.5*math.Sin(2*math.Pi*op.fHigh*t), 0)
			}

			data, err := common.Complex128SliceToBytes(signal)
			if err != nil {
				return
			}

			record := common2.Record{
				DataType: common2.DataType_BINARY,
				Data:     data,
			}

			op.InputChannel <- record

			count++
			time.Sleep(1 * time.Second)
		}
	}()
}

func (op *SimpleFFTSource) Compute([]byte) ([]byte, error) {
	return nil, nil
}

func (op *SimpleFFTSource) SetName(name string) {
	op.name = name
}

func (op *SimpleFFTSource) IsSourceOp() bool {
	return true
}

func (op *SimpleFFTSource) IsSinkOp() bool {
	return false
}

func (op *SimpleFFTSource) IsKeyByOp() bool {
	return false
}

func (op *SimpleFFTSource) Checkpoint() []byte {
	return nil
}

func (op *SimpleFFTSource) RestoreFromCheckpoint([]byte) error {
	return nil
}
