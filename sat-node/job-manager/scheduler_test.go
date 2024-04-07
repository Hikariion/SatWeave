package job_manager

//
//import (
//	"github.com/stretchr/testify/assert"
//	"satweave/sat-node/task-manager"
//	"satweave/utils/logger"
//	"testing"
//)
//
//func TestScheduler(t *testing.T) {
//	testScheduler(t)
//}
//
//func testScheduler(t *testing.T) {
//	yamlPath := "./test-files/task_config.yaml"
//	taskYaml := new(task_manager.TaskYaml)
//	taskYaml, err := task_manager.GetLogicalTaskFromYaml(yamlPath)
//	assert.NoError(t, err)
//	scheduler := new(UserDefinedScheduler)
//	executeTasks, err := scheduler.Schedule(taskYaml.Tasks)
//	assert.NoError(t, err)
//	logger.Infof("%v", executeTasks)
//}
