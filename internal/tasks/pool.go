package tasks

import (
	"time"

	"github.com/SwissOpenEM/globus"
	"github.com/SwissOpenEM/globus-transfer-service/internal/serviceuser"
	"github.com/alitto/pond/v2"
)

type TaskPool struct {
	scicatUrl         string
	globusClient      globus.GlobusClient
	scicatServiceUser serviceuser.ScicatServiceUser
	pool              pond.Pool
	taskPollInterval  time.Duration
}

func CreateTaskPool(scicatUrl string, globusClient globus.GlobusClient, scicatServiceUser serviceuser.ScicatServiceUser, maxConcurrency int, queueSize int, taskPollInterval uint) TaskPool {
	return TaskPool{
		scicatUrl:         scicatUrl,
		globusClient:      globusClient,
		scicatServiceUser: scicatServiceUser,
		pool:              pond.NewPool(maxConcurrency, pond.WithQueueSize(queueSize)),
		taskPollInterval:  time.Duration(taskPollInterval) * time.Second,
	}
}

func (tp TaskPool) AddTransferTask(globusTaskId string, datasetPid string, scicatJobId string) pond.Task {
	task := transferTask{
		scicatUrl:         &tp.scicatUrl,
		globusClient:      tp.globusClient,
		scicatServiceUser: tp.scicatServiceUser,
		globusTaskId:      globusTaskId,
		datasetPid:        datasetPid,
		scicatJobId:       scicatJobId,
		taskPollInterval:  tp.taskPollInterval,
	}

	return tp.pool.Submit(task.execute)
}

func (tp TaskPool) CanSubmitJob() bool {
	if tp.pool.QueueSize() == 0 {
		return true
	}
	return tp.pool.WaitingTasks() < uint64(tp.pool.QueueSize())
}

func (tp TaskPool) IsQueueSizeLimited() bool {
	return tp.pool.QueueSize() > 0
}
