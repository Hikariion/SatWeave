package sun

import "satweave/messenger/common"

type AcknowledgeTable struct {
	// checkpointId -> pending_checkpoint
	executeTaskMap     map[uint64][]*common.ExecuteTask
	pendingCheckpoints map[uint64]map[string]bool
}

func (a *AcknowledgeTable) hasCheckpoint(checkpointId uint64) bool {
	for key, _ := range a.pendingCheckpoints {
		if key == checkpointId {
			return true
		}
	}
	return false
}

func (a *AcknowledgeTable) registerPendingCheckpoint(checkpointId uint64) {

}
