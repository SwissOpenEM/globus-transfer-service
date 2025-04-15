package tasks

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SwissOpenEM/globus"
	"github.com/SwissOpenEM/globus-transfer-service/internal/serviceuser"
	"github.com/SwissOpenEM/globus-transfer-service/jobs"
	"github.com/paulscherrerinstitute/scicat-cli/v3/datasetIngestor"
)

type transferTask struct {
	scicatUrl         *string
	globusClient      globus.GlobusClient
	scicatServiceUser serviceuser.ScicatServiceUser
	globusTaskId      string
	datasetPid        string
	scicatJobId       string
	taskPollInterval  time.Duration
	cancel            chan struct{}

	// current status
	bytesTransferred uint
	filesTransferred uint
	filesTotal       uint
}

func (t transferTask) execute() {
	completed := false
	var err error = nil
	for {
		select {
		case <-t.cancel:
			t.cancelTask()
			return
		default:
		}
		completed, err = t.updateTask()
		if completed || err != nil {
			break
		}
		time.Sleep(t.taskPollInterval)
	}

	if !completed || err != nil {
		return // if not completed or error'd, don't mark the dataset as archivable
	}
	t.finishTask()
}

func (t transferTask) updateTask() (bool, error) {
	bytesTransferred, filesTransferred, totalFiles, completed, err := checkTransfer(t.globusClient, t.globusTaskId)

	status := jobs.Transferring
	statusCode := "002"
	statusMessage := "transferring"
	errMsg := ""

	if err != nil {
		status = jobs.Failed
		statusCode = "998"
		statusMessage = "an error has occured during task polling, this job is not updated anymore"
		errMsg = err.Error()
	} else if completed {
		status = jobs.Finished
		statusCode = "003"
		statusMessage = "finished"
	}

	taskLog(t.scicatJobId, t.globusTaskId, t.datasetPid, bytesTransferred, filesTransferred, totalFiles, status, err)

	token, err := t.scicatServiceUser.GetToken()
	if err != nil {
		errFull := fmt.Errorf("getting token failed, task with scicat job id '%s', dataset pid '%s', globus id '%s' cannot be updated: %s", t.scicatJobId, t.datasetPid, t.globusTaskId, err.Error())
		log.Println(errFull.Error())
		return false, errFull
	}

	_, err = UpdateGlobusTransferScicatJob(
		*t.scicatUrl,
		token,
		t.scicatJobId,
		statusCode,
		statusMessage,
		jobs.JobResultObject{
			GlobusTaskId:     t.globusTaskId,
			BytesTransferred: uint(bytesTransferred),
			FilesTransferred: uint(filesTransferred),
			FilesTotal:       uint(totalFiles),
			Status:           status,
			Error:            errMsg,
		},
	)

	return completed, err
}

func (t transferTask) finishTask() {
	token, _ := t.scicatServiceUser.GetToken()
	err := datasetIngestor.MarkFilesReady(http.DefaultClient, *t.scicatUrl+"api/v3", t.datasetPid, map[string]string{"accessToken": token})
	if err != nil {
		errMsg := err.Error()

		_, err = UpdateGlobusTransferScicatJob(
			*t.scicatUrl,
			token,
			t.scicatJobId,
			"997",
			"completed but can't mark dataset as archivable",
			jobs.JobResultObject{
				GlobusTaskId:     t.globusTaskId,
				BytesTransferred: t.bytesTransferred,
				FilesTransferred: t.filesTransferred,
				FilesTotal:       t.filesTotal,
				Status:           jobs.Finished,
				Error:            errMsg,
			},
		)
		if err != nil {
			taskLog(t.scicatJobId, t.globusTaskId, t.datasetPid, int(t.bytesTransferred), int(t.filesTransferred), int(t.filesTotal), jobs.Finished, err)
		}
	}
}

func (t transferTask) cancelTask() error {
	status := jobs.Cancelled
	statusCode := "003"
	statusMessage := "cancelled"
	errMsg := ""

	_, err := t.globusClient.TransferCancelTaskByID(t.globusTaskId)
	if err != nil {
		status = jobs.Failed
		statusCode = "996"
		statusMessage = "cancelling failed"
		errMsg = "failed cancelling globus transfer task: " + err.Error()
	}

	token, err := t.scicatServiceUser.GetToken()
	if err != nil {
		taskLog(t.scicatJobId, t.globusTaskId, t.datasetPid, int(t.bytesTransferred), int(t.filesTransferred), int(t.filesTotal), jobs.Cancelled, err)
		return err
	}

	_, err = UpdateGlobusTransferScicatJob(
		*t.scicatUrl,
		token,
		t.scicatJobId,
		statusCode,
		statusMessage,
		jobs.JobResultObject{
			GlobusTaskId:     t.globusTaskId,
			BytesTransferred: t.bytesTransferred,
			FilesTransferred: t.filesTransferred,
			FilesTotal:       t.filesTotal,
			Status:           status,
			Error:            errMsg,
		},
	)

	return err
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

func taskLog(sciacatJobId string, globusTaskId string, datasetPid string, bytesTransferred int, filesTransferred int, totalFiles int, status jobs.JobStatus, err error) {
	errString := ""
	if err != nil {
		errString = err.Error()
	}
	log.Printf("'%s' scicat job, '%s' globus task for '%s' dataset - bytes transferred: %d, files transferred: %d, total files detected: %d, status %s, error message: '%s'\n", sciacatJobId, globusTaskId, datasetPid, bytesTransferred, filesTransferred, totalFiles, status, errString)
}
