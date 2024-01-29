package sun

//import (
//	"satweave/messenger/common"
//	"satweave/utils/errno"
//	"satweave/utils/logger"
//	"sync"
//)
//
//type CheckpointCoordinator struct {
//	mutex                      sync.Mutex
//	RegisteredTaskManagerTable *RegisteredTaskManagerTable
//	table                      map[string]*SpecificJobInfo
//}
//
//func (c *CheckpointCoordinator) registerJob(jobId string, executeTaskMap map[uint64][]*common.ExecuteTask) error {
//	c.mutex.Lock()
//	defer c.mutex.Unlock()
//
//	if _, exists := c.table[jobId]; exists {
//		logger.Errorf("Failed to register job: jobId %v already exists", jobId)
//		return errno.RegisterJobFail
//	}
//
//	c.table[jobId] = NewSpecificJobInfo(executeTaskMap)
//
//	return nil
//}
//
//func (c *CheckpointCoordinator) triggerCheckpoint(jobId string, checkpointId int64, cancelJob bool) error {
//	c.mutex.Lock()
//	defer c.mutex.Unlock()
//
//	if _, exists := c.table[jobId]; !exists {
//		logger.Errorf("Failed to trigger checkpoint: jobId %v does not exist", jobId)
//		return errno.TriggerCheckpointFail
//	}
//
//	err := c.table[jobId].triggerCheckpoint(checkpointId, c.RegisteredTaskManagerTable, cancelJob)
//	if err != nil {
//		logger.Errorf("Failed to trigger checkpoint: %v", err)
//		return err
//	}
//
//	return nil
//}
//
//func (c *CheckpointCoordinator) AcknowledgeCheckpoint(request *AcknowledgeCheckpointRequest) (bool, error) {
//	c.mutex.Lock()
//	defer c.mutex.Unlock()
//
//	jobId := request.JobId
//	if _, exists := c.table[jobId]; !exists {
//		logger.Errorf("Failed to acknowledge checkpoint: jobId %v does not exist", jobId)
//		return false, errno.JobNotExist
//	}
//	return c.table[jobId].AcknowledgeCheckpoint(request)
//
//}
//
//func NewCheckpointCoordinator(registeredTaskManagerTable *RegisteredTaskManagerTable) *CheckpointCoordinator {
//	return &CheckpointCoordinator{
//		mutex:                      sync.Mutex{},
//		RegisteredTaskManagerTable: registeredTaskManagerTable,
//		table:                      make(map[string]*SpecificJobInfo),
//	}
//}
