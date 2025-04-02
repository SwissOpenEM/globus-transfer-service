package tasks

import (
	"time"

	"github.com/SwissOpenEM/globus"
	"github.com/alitto/pond/v2"
)

type TaskPool struct {
	scicatUrl        string
	globusClient     globus.GlobusClient
	pool             pond.Pool
	taskPollInterval time.Duration
}

func CreateTaskPool(scicatUrl string, globusClient globus.GlobusClient, maxConcurrency int, queueSize int, taskPollInterval uint) TaskPool {
	return TaskPool{
		scicatUrl:        scicatUrl,
		globusClient:     globusClient,
		pool:             pond.NewPool(maxConcurrency, pond.WithQueueSize(queueSize)),
		taskPollInterval: time.Duration(taskPollInterval) * time.Second,
	}
}

func (tp TaskPool) AddTransferTask(globusTaskId string, datasetPid string) pond.Task {
	task := transferTask{
		scicatUrl:        &tp.scicatUrl,
		globusClient:     tp.globusClient,
		globusTaskId:     globusTaskId,
		datasetPid:       datasetPid,
		taskPollInterval: tp.taskPollInterval,
	}

	return tp.pool.Submit(task.execute)
}
