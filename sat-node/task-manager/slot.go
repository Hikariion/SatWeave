package task_manager

import (
	"satweave/messenger/common"
	"satweave/sat-node/worker"
	"satweave/utils/logger"
)

type slotState uint64

const (
	deployed slotState = iota
	unDeployed
)

type Slot struct {
	raftID   uint64
	workerID uint64
	subTask  *worker.Worker
	status   slotState
}

func (s *Slot) start() {
	logger.Infof("raft id %v subtask %v begin to run...", s.raftID, s.subTask.SubTaskName)
	s.subTask.Run()
}

func NewSlot(raftId uint64, executeTask *common.ExecuteTask, jobManagerHost string, jobManagerPort uint64, jobId string) *Slot {
	return &Slot{
		raftID:   raftId,
		subTask:  worker.NewWorker(raftId, executeTask, jobManagerHost, jobManagerPort, jobId),
		status:   deployed,
		workerID: executeTask.WorkerId,
	}
}
