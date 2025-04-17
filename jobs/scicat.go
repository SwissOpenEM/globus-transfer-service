package jobs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Dataset struct {
	Pid   string   `json:"pid"`
	Files []string `json:"files"`
}

type JobParams struct {
	DatasetList []Dataset `json:"datasetList"`
}

type JobStatus string

const (
	Cancelled    JobStatus = "cancelled"
	Failed       JobStatus = "failed"
	Finished     JobStatus = "finished"
	Transferring JobStatus = "transferring"
)

type JobResultObject struct {
	GlobusTaskId     string    `json:"globusTaskId"`
	BytesTransferred uint      `json:"bytesTransferred"`
	FilesTransferred uint      `json:"filesTransferred"`
	FilesTotal       uint      `json:"filesTotal"`
	Status           JobStatus `json:"status"`
	Error            string    `json:"error"`
}

type ScicatJob struct {
	CreatedBy       string          `json:"createdBy"`
	UpdatedBy       string          `json:"updatedBy"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	OwnerGroup      string          `json:"ownerGroup"`
	AccessGroups    []string        `json:"accessGroups"`
	ID              string          `json:"id"`
	OwnerUser       string          `json:"ownerUser"`
	Type            string          `json:"type"`
	StatusCode      string          `json:"statusCode"`
	StatusMessage   string          `json:"statusMessage"`
	JobParams       JobParams       `json:"jobParams"`
	ContactEmail    string          `json:"contactEmail"`
	ConfigVersion   string          `json:"configVersion"`
	JobResultObject JobResultObject `json:"jobResultObject"`
}

type JobNotFoundErr struct {
	msg string
}

func (e *JobNotFoundErr) Error() string {
	return e.msg
}

func GetJobList(scicatUrl string, scicatToken string, filter string) ([]ScicatJob, error) {
	url, err := url.JoinPath(scicatUrl, "api", "v4", "jobs")
	if err != nil {
		return []ScicatJob{}, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []ScicatJob{}, err
	}

	q := req.URL.Query()
	q.Set("filter", filter)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+scicatToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []ScicatJob{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		return []ScicatJob{}, fmt.Errorf("getting transfer job list failed: bad request - likely bad filter was passed")
	}
	if resp.StatusCode != 200 {
		return []ScicatJob{}, fmt.Errorf("getting transfer job list failed with unknwon error - status code %d, status %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []ScicatJob{}, err
	}

	jobs := []ScicatJob{}
	err = json.Unmarshal(body, &jobs)
	return jobs, err
}

func GetJobById(scicatUrl string, scicatToken string, jobId string) (ScicatJob, error) {
	url, err := url.JoinPath(scicatUrl, "api", "v4", "jobs", jobId)
	if err != nil {
		return ScicatJob{}, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ScicatJob{}, err
	}

	req.Header.Set("Authorization", "Bearer "+scicatToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ScicatJob{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return ScicatJob{}, fmt.Errorf("user doesn't have the right to access this dataset or user credentials are invalid")
	}
	if resp.StatusCode == 400 {
		b, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(b), "Invalid job id") {
			return ScicatJob{}, &JobNotFoundErr{"Job not found"}
		}
	}
	if resp.StatusCode != 200 {
		return ScicatJob{}, fmt.Errorf("unknown error - statuscode: %d, status: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScicatJob{}, err
	}

	job := ScicatJob{}
	err = json.Unmarshal(body, &job)
	return job, err
}
