package sun

import (
	"errors"
	"satweave/messenger/common"
	"satweave/utils/logger"
	"sync"
)

type RegisteredTaskManagerTable struct {
	mutex sync.Mutex
	table map[uint64]*common.TaskManagerDescription
}

func (r *RegisteredTaskManagerTable) register(description *common.TaskManagerDescription) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, exists := r.table[description.RaftId]; exists {
		logger.Errorf("TaskManager %v already exists", description.RaftId)
		return errors.New("TaskManager  already exists")
	}
	r.table[description.RaftId] = description
	logger.Infof("Register TaskManager %v success", description.RaftId)
	return nil
}

func (r *RegisteredTaskManagerTable) hasTaskManager(raftId uint64) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, exists := r.table[raftId]
	return exists
}

func (r *RegisteredTaskManagerTable) getHost(raftId uint64) string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.table[raftId].Host
}

func (r *RegisteredTaskManagerTable) getPort(raftId uint64) uint64 {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.table[raftId].Port
}

func newRegisteredTaskManagerTable() *RegisteredTaskManagerTable {
	return &RegisteredTaskManagerTable{
		table: make(map[uint64]*common.TaskManagerDescription),
		mutex: sync.Mutex{},
	}
}
