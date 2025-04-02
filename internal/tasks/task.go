package tasks

import (
	"fmt"
	"log"
	"time"

	"github.com/SwissOpenEM/globus"
)

type transferTask struct {
	scicatUrl        *string
	globusClient     globus.GlobusClient
	globusTaskId     string
	datasetPid       string
	taskPollInterval time.Duration
}

func (t transferTask) execute() {
	var bytesTransferred, filesTransferred, totalFiles int
	var completed bool
	var err error
	for {
		bytesTransferred, filesTransferred, totalFiles, completed, err = checkTransfer(t.globusClient, t.globusTaskId)
		updateTaskInScicat(t.globusTaskId, t.datasetPid, bytesTransferred, filesTransferred, totalFiles, completed, err)
		if completed || (err != nil) {
			break
		}
		time.Sleep(t.taskPollInterval)
	}
}

func checkTransfer(client globus.GlobusClient, globusTaskId string) (bytesTransferred int, filesTransferred int, totalFiles int, completed bool, err error) {
	globusTask, err := client.TransferGetTaskByID(globusTaskId)
	if err != nil {
		return 0, 0, 1, false, fmt.Errorf("globus: can't continue transfer because an error occured while polling the task \"%s\": %v", globusTaskId, err)
	}
	switch globusTask.Status {
	case "ACTIVE":
		totalFiles := globusTask.Files
		if globusTask.FilesSkipped != nil {
			totalFiles -= *globusTask.FilesSkipped
		}
		return globusTask.BytesTransferred, globusTask.FilesTransferred, totalFiles, false, nil
	case "INACTIVE":
		return 0, 0, 1, false, fmt.Errorf("globus: transfer became inactive, manual intervention required")
	case "SUCCEEDED":
		totalFiles := globusTask.Files
		if globusTask.FilesSkipped != nil {
			totalFiles -= *globusTask.FilesSkipped
		}
		return globusTask.BytesTransferred, globusTask.FilesTransferred, totalFiles, true, nil
	case "FAILED":
		return 0, 0, 1, false, fmt.Errorf("globus: task failed with the following error - code: \"%s\" description: \"%s\"", globusTask.FatalError.Code, globusTask.FatalError.Description)
	default:
		return 0, 0, 1, false, fmt.Errorf("globus: unknown task status: %s", globusTask.Status)
	}
}

func updateTaskInScicat(globusTaskId string, datasetPid string, bytesTransferred int, filesTransferred int, totalFiles int, completed bool, err error) {
	log.Printf("'%s' task for '%s' dataset - bytes transferred: %d, files transferred: %d, total files detected: %d, completed %v, error message: '%s'\n", globusTaskId, datasetPid, bytesTransferred, filesTransferred, totalFiles, completed, err.Error())
}
