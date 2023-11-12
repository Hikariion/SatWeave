package operators

// OperatorBase
// 用户自定义算子的基类
type OperatorBase interface {
	SetName(name string)
	Init(map[string]string)
	Compute(data []byte) ([]byte, error)
	IsSourceOp() bool
	IsSinkOp() bool
	IsKeyByOp() bool
}
