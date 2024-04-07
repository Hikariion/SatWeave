package generator

import "sync"

// CheckpointIdGenerator 结构体
type CheckpointIdGenerator struct {
	counter int64
}

// 单例实例及其初始化控制
var (
	checkpointIdGeneratorInstance *CheckpointIdGenerator
	checkpointIdGeneratorOnce     sync.Once
)

// GetCheckpointIdGeneratorInstance 返回 CheckpointIdGenerator 的单例实例
func GetCheckpointIdGeneratorInstance() *CheckpointIdGenerator {
	checkpointIdGeneratorOnce.Do(func() {
		checkpointIdGeneratorInstance = &CheckpointIdGenerator{
			counter: 0,
		}
	})
	return checkpointIdGeneratorInstance
}

// Next 返回下一个ID
func (c *CheckpointIdGenerator) Next() int64 {
	c.counter++
	return c.counter
}
