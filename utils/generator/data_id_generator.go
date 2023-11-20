package generator

import "sync"

// DataIdGenerator 结构体
type DataIdGenerator struct {
	counter uint64
}

// 单例实例及其初始化控制
var (
	dataIdGeneratorInstance *DataIdGenerator
	dataIdGeneratorOnce     sync.Once
)

// GetDataIdGeneratorInstance 返回 DataIdGenerator 的单例实例
func GetDataIdGeneratorInstance() *DataIdGenerator {
	dataIdGeneratorOnce.Do(func() {
		dataIdGeneratorInstance = &DataIdGenerator{
			counter: 0,
		}
	})
	return dataIdGeneratorInstance
}

// Next 返回下一个ID
func (d *DataIdGenerator) Next() uint64 {
	d.counter++
	return d.counter
}
