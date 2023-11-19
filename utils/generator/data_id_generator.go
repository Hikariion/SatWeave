package generator

import "sync"

// DataIdGenerator 结构体
type DataIdGenerator struct {
	counter int64
}

// 单例实例及其初始化控制
var (
	instance *DataIdGenerator
	once     sync.Once
)

// GetDataIdGeneratorInstance 返回 DataIdGenerator 的单例实例
func GetDataIdGeneratorInstance() *DataIdGenerator {
	once.Do(func() {
		instance = &DataIdGenerator{
			counter: 0,
		}
	})
	return instance
}

// Next 返回下一个ID
func (d *DataIdGenerator) Next() int64 {
	d.counter++
	return d.counter
}
