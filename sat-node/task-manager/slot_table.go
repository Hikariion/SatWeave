package task_manager

import (
	"satweave/messenger/common"
	"satweave/sat-node/worker"
	"satweave/utils/errno"
	"satweave/utils/logger"
	"sync"
)

type SlotTable struct {
	satelliteName string
	capacity      uint64
	table         map[string]*Slot // subtask_name -> slot

	mutex *sync.Mutex

	jobManagerHost string
	jobManagerPort uint64
}

func (st *SlotTable) getWorkerByWorkerId(id uint64) (*worker.Worker, error) {
	for _, slot := range st.table {
		if slot.workerID == id {
			return slot.subTask, nil
		}
	}
	return nil, errno.WorkerNotFound
}

func (st *SlotTable) deployExecuteTask(jobId string, executeTask *common.ExecuteTask,
	pathNodes []string, yamlBytes []byte) error {
	/*
		add a slot by execute_task
	*/
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if len(st.table) >= int(st.capacity) {
		logger.Errorf("satellite  %s slot table length >= capacity", st.satelliteName)
		return errno.SlotCapacityNotEnough
	}

	subtaskName := executeTask.SubtaskName
	if _, ok := st.table[subtaskName]; ok {
		logger.Errorf("satellite %s slot table has subtask %s", st.satelliteName, subtaskName)
		return errno.RequestSlotFail
	}

	st.table[subtaskName] = NewSlot(st.satelliteName, executeTask, st.jobManagerHost, st.jobManagerPort, jobId, pathNodes, yamlBytes)

	logger.Infof("satellite %s deploy subtask %s success", st.satelliteName, subtaskName)

	return nil
}

func (st *SlotTable) startExecuteTask(subtaskName string) {
	slot := st.getSlot(subtaskName)
	slot.start()
	logger.Infof("satellite %s start subtask %s success", st.satelliteName, subtaskName)
}

func (st *SlotTable) hasSlot(name string) bool {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	_, ok := st.table[name]
	return ok
}

func (st *SlotTable) getSlot(name string) *Slot {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	return st.table[name]
}

func NewSlotTable(satelliteName string, capacity uint64, jobManagerHost string, jobManagerPort uint64) *SlotTable {
	return &SlotTable{
		satelliteName: satelliteName,
		capacity:      capacity,
		table:         make(map[string]*Slot),
		mutex:         &sync.Mutex{},

		jobManagerHost: jobManagerHost,
		jobManagerPort: jobManagerPort,
	}
}
