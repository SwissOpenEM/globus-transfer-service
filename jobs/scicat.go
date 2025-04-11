package jobs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Dataset struct {
	Pid   string   `json:"pid"`
	Files []string `json:"files"`
}

type JobParams struct {
	DatasetList []Dataset `json:"datasetList"`
}

type JobResultObject struct {
	GlobusTaskId     string `json:"globusTaskId"`
	BytesTransferred uint   `json:"bytesTransferred"`
	FilesTransferred uint   `json:"filesTransferred"`
	FilesTotal       uint   `json:"filesTotal"`
	Completed        bool   `json:"completed"`
	Error            string `json:"error"`
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScicatJob{}, err
	}

	job := ScicatJob{}
	err = json.Unmarshal(body, &job)
	return job, err
}
