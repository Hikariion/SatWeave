package task_manager

import (
	"satweave/messenger/common"
	"satweave/sat-node/worker"
	"satweave/utils/errno"
	"satweave/utils/logger"
	"sync"
)

type SlotTable struct {
	raftId   uint64
	capacity uint64
	table    map[string]*Slot // subtask_name -> slot

	mutex *sync.Mutex
}

func (st *SlotTable) getWorkerByWorkerId(id uint64) (*worker.Worker, error) {
	for _, slot := range st.table {
		if slot.workerID == id {
			return slot.subTask, nil
		}
	}
	return nil, errno.WorkerNotFound
}

func (st *SlotTable) deployExecuteTask(executeTask *common.ExecuteTask) error {
	/*
		add a slot by execute_task
	*/
	st.mutex.Lock()
	defer st.mutex.Unlock()

	if len(st.table) >= int(st.capacity) {
		logger.Errorf("raft %d slot table length >= capacity", st.raftId)
		return errno.SlotCapacityNotEnough
	}

	subtaskName := executeTask.SubtaskName
	if _, ok := st.table[subtaskName]; ok {
		logger.Errorf("raft %d slot table has subtask %s", st.raftId, subtaskName)
		return errno.RequestSlotFail
	}

	st.table[subtaskName] = NewSlot(st.raftId, executeTask)

	logger.Infof("raft %d deploy subtask %s success", st.raftId, subtaskName)

	return nil
}

func (st *SlotTable) startExecuteTask(subtaskName string) {
	slot := st.getSlot(subtaskName)
	slot.start()
	logger.Infof("raft %d start subtask %s success", st.raftId, subtaskName)
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

func NewSlotTable(raftId uint64, capacity uint64) *SlotTable {
	return &SlotTable{
		raftId:   raftId,
		capacity: capacity,
		table:    make(map[string]*Slot),
		mutex:    &sync.Mutex{},
	}
}
