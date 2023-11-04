package job_manager

import (
	"context"
	"satweave/messenger/common"
	"satweave/sat-node/task-manager"
	"satweave/sat-node/watcher"
	"satweave/utils/logger"
)

type JobManager struct {
	UnimplementedJobManagerServiceServer
	scheduler UserDefinedScheduler
	watcher   *watcher.Watcher
}

func (j *JobManager) SubmitJob(ctx context.Context, request *SubmitJobRequest) (*common.NilResponse, error) {
	tasks := request.Tasks
	err := j.innerSubmitJob(ctx, tasks)
	if err != nil {
		logger.Errorf("Error submitting job: %s\n", err)
		return nil, err
	}
	return &common.NilResponse{}, nil
}

func (j *JobManager) innerSubmitJob(ctx context.Context, tasks []*task_manager.Task) error {
	// schedule
	executeMap, err := j.scheduler.Schedule(0, tasks)
	if err != nil {
		logger.Errorf("Error scheduling tasks: %s\n", err)
		return err
	}

	// deploy
	err = j.deployExecuteTasks(ctx, executeMap)
	if err != nil {
		logger.Errorf("Error deploying tasks: %s\n", err)
		return err
	}

	// start
	return nil
}

// TODO(qiu): 发布一个 ExecuteTask 的 info？
func (j *JobManager) deployExecuteTasks(ctx context.Context, executeMap map[string][]*task_manager.ExecuteTask) error {

	return nil
}

func (j *JobManager) startExecuteTasks() {

}
