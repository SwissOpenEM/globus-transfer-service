package tasks

import "github.com/alitto/pond/v2"

type TaskPool struct {
	scicatUrl string
	pool      pond.Pool
}

func CreateTaskPool(scicatUrl string, maxConcurrency int, queueSize int) TaskPool {
	return TaskPool{
		scicatUrl: scicatUrl,
		pool:      pond.NewPool(maxConcurrency, pond.WithQueueSize(queueSize)),
	}
}

func (tp TaskPool) AddTransferTask(globusTaskId string, datasetPid string) pond.Task {
	task := transferTask{
		scicatUrl:    &tp.scicatUrl,
		globusTaskId: globusTaskId,
		datasetPid:   datasetPid,
	}

	return tp.pool.Submit(task.execute)
}
