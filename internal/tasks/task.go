package tasks

import (
	"fmt"
	"log"
	"time"

	"github.com/SwissOpenEM/globus"
	"github.com/SwissOpenEM/globus-transfer-service/internal/serviceuser"
)

type transferTask struct {
	scicatUrl         *string
	globusClient      globus.GlobusClient
	scicatServiceUser serviceuser.ScicatServiceUser
	globusTaskId      string
	datasetPid        string
	scicatJobId       string
	taskPollInterval  time.Duration
}

func (t transferTask) execute() {
	var bytesTransferred, filesTransferred, totalFiles int
	var completed bool
	var err error
	for {
		bytesTransferred, filesTransferred, totalFiles, completed, err = checkTransfer(t.globusClient, t.globusTaskId)
		taskLog(t.scicatJobId, t.globusTaskId, t.datasetPid, bytesTransferred, filesTransferred, totalFiles, completed, err)
		statusCode := "002"
		statusMessage := "transferring"
		errMsg := ""
		if err != nil {
			statusCode = "998"
			statusMessage = "an error has occured during task polling, this job is not updated anymore"
			errMsg = err.Error()
		}
		token, err := t.scicatServiceUser.GetToken()
		if err != nil {
			log.Fatalf("getting token failed, task with scicat job id '%s', dataset pid '%s', globus id '%s' cannot be updated: %s", t.scicatJobId, t.datasetPid, t.globusTaskId, err.Error())
		}
		if completed {
			statusCode = "003"
			statusMessage = "finished"
		}
		UpdateGlobusTransferScicatJob(
			*t.scicatUrl,
			token,
			t.scicatJobId,
			statusCode,
			statusMessage, GlobusTransferScicatJobResultObject{
				BytesTransferred: uint(bytesTransferred),
				FilesTransferred: uint(filesTransferred),
				FilesTotal:       uint(totalFiles),
				Completed:        completed,
				Error:            errMsg,
			},
		)
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

func taskLog(sciacatJobId string, globusTaskId string, datasetPid string, bytesTransferred int, filesTransferred int, totalFiles int, completed bool, err error) {
	errString := ""
	if err != nil {
		errString = err.Error()
	}
	log.Printf("'%s' scicat job, '%s' globus task for '%s' dataset - bytes transferred: %d, files transferred: %d, total files detected: %d, completed %v, error message: '%s'\n", sciacatJobId, globusTaskId, datasetPid, bytesTransferred, filesTransferred, totalFiles, completed, errString)
}
