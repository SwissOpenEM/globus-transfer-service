package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
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
	BytesTransferred uint   `json:"bytesTransferred,omitempty"`
	FilesTransferred uint   `json:"filesTransferred,omitempty"`
	FilesTotal       uint   `json:"filesTotal,omitempty"`
	Completed        bool   `json:"completed,omitempty"`
	Error            string `json:"error,omitempty"`
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

func CreateGlobusTransferScicatJob(scicatUrl string, scicatToken string, ownerGroup string, datasetPid string) (GlobusTransferScicatJob, error) {
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
