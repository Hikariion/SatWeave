package operators

import (
	"math"
	"math/rand"
	common2 "satweave/messenger/common"
	"satweave/utils/common"
	"time"
)

type SimpleFFTSource struct {
	name string
	// 用于 Source 算子和 Worker 之间通信
	InputChannel chan *common2.Record

	// For dsp
	fs    float64
	T     float64
	fLow  float64
	fHigh float64
}

func (op *SimpleFFTSource) Init(initMap map[string]interface{}) {
	op.InputChannel = initMap["InputChannel"].(chan *common2.Record)

	op.fs = 1000.0
	op.T = 1.0
	op.fHigh = 50.0
	op.fLow = 5.0

	go func() {
		count := 0
		for {
			if count >= 10 {
				break
			}
			N := int(op.fs * op.T)

			rand.Seed(time.Now().UnixNano()) // 初始化随机数种子

			signal := make([]complex128, N)
			// 为每次循环生成随机的频率和相位
			fLow := rand.Float64()*5 + 5              // 产生5到10之间的随机低频
			fHigh := rand.Float64()*45 + 5            // 产生5到50之间的随机高频
			phaseLow := rand.Float64() * 2 * math.Pi  // 为低频生成随机相位
			phaseHigh := rand.Float64() * 2 * math.Pi // 为高频生成随机相位
			for n := 0; n < N; n++ {
				if rand.Float64() <= 2.0/3.0 {
					t := float64(n) / op.fs
					signal[n] = complex(math.Sin(2*math.Pi*fLow*t+phaseLow)+0.5*math.Sin(2*math.Pi*fHigh*t+phaseHigh), 0)
				}
			}

			data, err := common.Complex128SliceToBytes(signal)
			if err != nil {
				return
			}

			record := &common2.Record{
				DataType: common2.DataType_BINARY,
				Data:     data,
			}

			op.InputChannel <- record

			count++

			time.Sleep(1 * time.Second)
		}

		record := &common2.Record{
			DataType: common2.DataType_FINISH,
			Data:     nil,
		}

		op.InputChannel <- record

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
