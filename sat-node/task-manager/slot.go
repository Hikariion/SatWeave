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

// Slot 对应一个 subtask
type Slot struct {
	satelliteName string
	workerID      uint64
	subTask       *worker.Worker
	status        slotState
	jobId         string

	jobManagerHost string
	jobManagerPort uint64
}

func (s *Slot) start() {
	logger.Infof("satellite name  %v subtask %v begin to run...", s.satelliteName, s.subTask.SubTaskName)
	s.subTask.Run()
}

func NewSlot(satelliteName string, executeTask *common.ExecuteTask, jobManagerHost string, jobManagerPort uint64, jobId string) *Slot {
	return &Slot{
		satelliteName:  satelliteName,
		subTask:        worker.NewWorker(satelliteName, executeTask, jobManagerHost, jobManagerPort, jobId),
		status:         deployed,
		workerID:       executeTask.WorkerId,
		jobId:          jobId,
		jobManagerHost: jobManagerHost,
		jobManagerPort: jobManagerPort,
	}
}
