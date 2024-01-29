package sun

import (
	"errors"
	"satweave/messenger/common"
	"satweave/utils/logger"
	"sync"
)

type RegisteredTaskManagerTable struct {
	mutex sync.Mutex
	table map[string]*common.TaskManagerDescription
}

func (r *RegisteredTaskManagerTable) register(description *common.TaskManagerDescription) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if t, exists := r.table[description.SatelliteName]; exists {
		if t.HostPort.GetHost() == description.HostPort.GetHost() &&
			t.HostPort.GetPort() == description.HostPort.GetPort() {
			logger.Errorf("TaskManager %v already exists", description.SatelliteName)
			return errors.New("TaskManager  already exists")
		}
	}
	r.table[description.SatelliteName] = description
	logger.Infof("Register TaskManager %v success", description.SatelliteName)
	return nil
}

func (r *RegisteredTaskManagerTable) hasTaskManager(SatelliteName string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, exists := r.table[SatelliteName]
	return exists
}

func (r *RegisteredTaskManagerTable) getHostPort(SatelliteName string) *common.HostPort {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.table[SatelliteName].HostPort
}

func newRegisteredTaskManagerTable() *RegisteredTaskManagerTable {
	return &RegisteredTaskManagerTable{
		table: make(map[string]*common.TaskManagerDescription),
		mutex: sync.Mutex{},
	}
}
