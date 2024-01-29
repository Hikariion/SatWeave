package generator

import (
	"github.com/google/uuid"
	"sync"
)

// JobIdGenerator 结构体
type JobIdGenerator struct {
}

// 单例实例及其初始化控制
var (
	jobIdGeneratorInstance *JobIdGenerator
	jobIdGeneratorOnce     sync.Once
)

// GetJobIdGeneratorInstance 返回 JobIdGenerator 的单例实例
func GetJobIdGeneratorInstance() *JobIdGenerator {
	jobIdGeneratorOnce.Do(func() {
		jobIdGeneratorInstance = &JobIdGenerator{}
	})
	return jobIdGeneratorInstance
}

// Next 返回下一个ID
func (d *JobIdGenerator) Next() string {
	uid := uuid.New()
	return uid.String()
}
