package operators

import "satweave/messenger/common"

type OperatorBase interface {
	SetName(name string)
	Init(map[string]string)
	Compute(record *common.Record)
}
