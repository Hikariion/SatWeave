package operators

// OperatorBase
// 用户自定义算子的基类
type OperatorBase interface {
	SetName(name string)
	Init(map[string]string)
	Compute(data []byte) ([]byte, error)
	Checkpoint() []byte // 返回值为 snapshot 状态，作为文件存储 common_pb.File
	IsSourceOp() bool
	IsSinkOp() bool
	IsKeyByOp() bool
	RestoreFromCheckpoint([]byte) error
}
