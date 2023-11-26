package sun

import (
	"github.com/scylladb/go-set/strset"
	"satweave/messenger/common"
	"satweave/utils/errno"
	"satweave/utils/logger"
)

type AcknowledgeTable struct {
	// checkpointId -> pending_checkpoint
	executeTaskMap     map[uint64][]*common.ExecuteTask
	pendingCheckpoints map[int64]*strset.Set // checkpointId -> subtaskName bool
}

func (a *AcknowledgeTable) hasCheckpoint(checkpointId int64) bool {
	if _, exists := a.pendingCheckpoints[checkpointId]; exists {
		return true
	}
	return false
}

func (a *AcknowledgeTable) registerPendingCheckpoint(checkpointId int64) {
	if _, exists := a.pendingCheckpoints[checkpointId]; exists {
		logger.Errorf("Failed to register pending checkpoint: checkpointId %v already exists", checkpointId)
		return
	}

	a.pendingCheckpoints[checkpointId] = strset.New()
	for _, tasks := range a.executeTaskMap {
		for _, task := range tasks {
			a.pendingCheckpoints[checkpointId].Add(task.SubtaskName)
		}
	}
}

func (a *AcknowledgeTable) acknowledgeCheckpoint(request *AcknowledgeCheckpointRequest) (bool, error) {
	// 返回是否完全 ack
	checkpointId := request.CheckpointId
	if _, exists := a.pendingCheckpoints[checkpointId]; !exists {
		logger.Errorf("Failed to acknowledge checkpoint: checkpointId %v does not exist", checkpointId)
		return false, errno.CheckpointIdNotInPending
	}

	subtaskName := request.SubtaskName
	pendingCheckpoint := a.pendingCheckpoints[checkpointId]
	if !pendingCheckpoint.Has(subtaskName) {
		logger.Warningf("%s already acknowledge", subtaskName)
	} else {
		pendingCheckpoint.Remove(subtaskName)
	}

	if pendingCheckpoint.Size() == 0 {
		delete(a.pendingCheckpoints, checkpointId)
		return true, nil
	}

	return false, nil
}

func NewAcknowledgeTable(executeTaskMap map[uint64][]*common.ExecuteTask) *AcknowledgeTable {
	return &AcknowledgeTable{
		executeTaskMap:     executeTaskMap,
		pendingCheckpoints: make(map[int64]*strset.Set),
	}
}
