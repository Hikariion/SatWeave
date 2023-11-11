package task_manager

import (
	"satweave/messenger/common"
	"satweave/sat-node/worker"
)

type slotState uint64

const (
	deployed slotState = iota
	unDeployed
)

type Slot struct {
	raftID  uint64
	subTask *worker.Worker
	status  slotState
}

func (s *Slot) start() {

}

func NewSlot(raftId uint64, executeTask *common.ExecuteTask) *Slot {
	return &Slot{
		raftID:  raftId,
		subTask: worker.NewWorker(raftId, executeTask),
		status:  deployed,
	}
}
