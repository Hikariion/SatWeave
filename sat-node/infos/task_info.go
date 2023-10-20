package infos

func (m *TaskInfo) GetInfoType() InfoType {
	return InfoType_TASK_INFO
}

func (m *TaskInfo) BaseInfo() *BaseInfo {
	return &BaseInfo{Info: &BaseInfo_TaskInfo{TaskInfo: m}}
}

func (m *TaskInfo) GetID() string {
	return m.TaskId
}

func (m *TaskInfo) GetSelfTask() []*TaskInfo {
	var tasks []*TaskInfo
	// TODO(qiu): here to get self task to process
	return tasks
}
