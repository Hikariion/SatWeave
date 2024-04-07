package operators

import (
	"context"
	"math/cmplx"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/utils/common"
	"satweave/utils/logger"
)

// LaSumOp 是低通聚合算子
type LaSumOp struct {
	JobId   string
	counter uint64
}

func (op *LaSumOp) SetJobId(JobId string) {
	op.JobId = JobId
}

func (op *LaSumOp) Init(initMap map[string]interface{}) {
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
	data := common.Uint64ToBytes(op.counter)
	return data
}

func (op *LaSumOp) RestoreFromCheckpoint(SunIp, ClsName string, SunPort uint64) error {
	conn, err := messenger.GetRpcConn(SunIp, SunPort)
	if err != nil {
		logger.Errorf("Fail to get rpc conn on TaskManager %v", SunIp)
		return err
	}
	client := sun.NewSunClient(conn)
	result, err := client.RestoreFromCheckpoint(context.Background(),
		&sun.RestoreFromCheckpointRequest{
			SubtaskName: ClsName,
		})
	if err != nil {
		return err
	}
	state := result.State
	if state == nil {
		op.counter = 0
	} else {
		op.counter = common.BytesToUint64(state)
	}
	return nil
}
