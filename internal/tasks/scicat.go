package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/SwissOpenEM/globus-transfer-service/internal/serviceuser"
)

type ScicatJobDatasetElement struct {
	Pid   string   `json:"pid"`
	Files []string `json:"files"`
}

type ScicatJobParams struct {
	DatasetList []ScicatJobDatasetElement `json:"datasetList"`
}

type scicatJobPost struct {
	Type         string          `json:"type"`
	JobParams    ScicatJobParams `json:"jobParams"`
	OwnerUser    string          `json:"ownerUser,omitempty"`
	OwnerGroup   string          `json:"ownerGroup,omitempty"`
	ContactEmail string          `json:"contactEmail,omitempty"`
}

type scicatGlobusTransferJobPatch struct {
	StatusCode      string                              `json:"statusCode,omitempty"`
	StatusMessage   string                              `json:"statusMessage,omitempty"`
	JobResultObject GlobusTransferScicatJobResultObject `json:"jobResultObject,omitempty"`
}

type GlobusTransferScicatJobResultObject struct {
	GlobusTaskId     string `json:"globusTaskId"`
	BytesTransferred uint   `json:"bytesTransferred"`
	FilesTransferred uint   `json:"filesTransferred"`
	FilesTotal       uint   `json:"filesTotal"`
	Completed        bool   `json:"completed"`
	Error            string `json:"error"`
}

type GlobusTransferScicatJob struct {
	CreatedBy       string                              `json:"createdBy"`
	UpdatedBy       string                              `json:"updatedBy"`
	CreatedAt       time.Time                           `json:"createdAt"`
	UpdatedAt       time.Time                           `json:"updatedAt"`
	OwnerGroup      string                              `json:"ownerGroup"`
	AccessGroups    []string                            `json:"accessGroups"`
	ID              string                              `json:"id"`
	OwnerUser       string                              `json:"ownerUser"`
	Type            string                              `json:"type"`
	StatusCode      string                              `json:"statusCode"`
	StatusMessage   string                              `json:"statusMessage"`
	JobParams       ScicatJobParams                     `json:"jobParams"`
	ContactEmail    string                              `json:"contactEmail"`
	ConfigVersion   string                              `json:"configVersion"`
	JobResultObject GlobusTransferScicatJobResultObject `json:"jobResultObject"`
}

func CreateGlobusTransferScicatJob(scicatUrl string, scicatToken string, ownerGroup string, datasetPid string, globusTaskId string) (GlobusTransferScicatJob, error) {
	url, err := url.JoinPath(scicatUrl, "api", "v4", "jobs")
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}

	reqBody, err := json.Marshal(scicatJobPost{
		Type:       "globus_transfer_job",
		OwnerGroup: ownerGroup,
		JobParams: ScicatJobParams{
			[]ScicatJobDatasetElement{
				{
					Pid:   datasetPid,
					Files: []string{},
				},
			},
		},
	})
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+scicatToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return GlobusTransferScicatJob{}, fmt.Errorf("authentication failed: user is not logged in")
	}
	if resp.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return GlobusTransferScicatJob{}, fmt.Errorf("unknown error occured with status: '%s', body: '%s'", resp.Status, string(bodyBytes))
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}

	job := GlobusTransferScicatJob{}
	err = json.Unmarshal(respBodyBytes, &job)
	if err != nil {
		return job, err
	}

	return UpdateGlobusTransferScicatJob(scicatUrl, scicatToken, job.ID, "001", "started", GlobusTransferScicatJobResultObject{
		GlobusTaskId:     globusTaskId,
		BytesTransferred: 0,
		FilesTransferred: 0,
		FilesTotal:       0,
		Completed:        false,
		Error:            "",
	})
}

func UpdateGlobusTransferScicatJob(scicatUrl string, scicatToken string, jobId string, statusCode string, statusMessage string, jobStatus GlobusTransferScicatJobResultObject) (GlobusTransferScicatJob, error) {
	url, err := url.JoinPath(scicatUrl, "api", "v4", "jobs", url.QueryEscape(jobId))
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}

	reqBody, err := json.Marshal(scicatGlobusTransferJobPatch{
		StatusCode:      statusCode,
		StatusMessage:   statusMessage,
		JobResultObject: jobStatus,
	})
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+scicatToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 403:
		return GlobusTransferScicatJob{}, fmt.Errorf("cannot patch dataset: forbidden")
	case 400:
		return GlobusTransferScicatJob{}, fmt.Errorf("cannot patch dataset: invalid job id")
	case 200:
		break
	default:
		body, _ := io.ReadAll(resp.Body)
		return GlobusTransferScicatJob{}, fmt.Errorf("unknown status encountered: '%s', body: '%s'", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GlobusTransferScicatJob{}, err
	}

	job := GlobusTransferScicatJob{}
	err = json.Unmarshal(body, &job)
	return job, err
}

func RestoreGlobusTransferJobsFromScicat(scicatUrl string, serviceUser serviceuser.ScicatServiceUser, pool TaskPool) error {
	url, err := url.JoinPath(scicatUrl, "api", "v4", "jobs")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// TODO: maybe add pagination or a limit to the filter
	q := req.URL.Query()
	q.Set("filter", `{"where":{"type":"globus_transfer_job","jobResultObject.completed":false,"jobResultObject.error":""}}`)
	req.URL.RawQuery = q.Encode()

	token, err := serviceUser.GetToken()
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	unfinishedJobs := []GlobusTransferScicatJob{}
	err = json.Unmarshal(body, &unfinishedJobs)
	if err != nil {
		return err
	}

	for _, job := range unfinishedJobs {
		if job.JobResultObject.GlobusTaskId == "" {
			log.Printf("Warning: job with id '%s' has no globus task id, so it cannot be resumed\n", job.ID)
			continue
		}
		if len(job.JobParams.DatasetList) > 1 {
			log.Printf("Warning: job with id '%s' has more than one associated dataset (a total of %d), which is not currently supported\n", job.ID, len(job.JobParams.DatasetList))
			continue
		}
		if len(job.JobParams.DatasetList) <= 0 {
			log.Printf("Warning: job with id '%s' has no datasets associated, so it cannot be resumed\n", job.ID)
		}
		pool.AddTransferTask(job.JobResultObject.GlobusTaskId, job.JobParams.DatasetList[0].Pid, job.ID)
	}

	return nil
}
