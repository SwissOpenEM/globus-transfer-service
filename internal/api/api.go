package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"slices"
	"text/template"

	"github.com/SwissOpenEM/globus"
	"github.com/SwissOpenEM/globus-transfer-service/internal/tasks"
	"github.com/gin-gonic/gin"
)

//go:generate oapi-codegen --config=cfg.yaml openapi.yaml

type ServerHandler struct {
	globusClient          globus.GlobusClient
	scicatUrl             string
	facilityCollectionIDs map[string]string
	srcGroupTemplate      *template.Template
	dstGroupTemplate      *template.Template
	dstPathTemplate       DestinationTemplate
	taskPool              tasks.TaskPool
}

type ScicatDataset struct {
	OwnerGroup   string `json:"ownerGroup"`
	SourceFolder string `json:"sourceFolder"`
}

var _ StrictServerInterface = ServerHandler{}

func NewServerHandler(globusClient globus.GlobusClient, scopes []string, scicatUrl string, facilityCollectionIDs map[string]string, srcGroupTemplateBody string, dstGroupTemplateBody string, dstPathTemplateBody string, taskPool tasks.TaskPool) (ServerHandler, error) {
	// create server with service client
	var err error
	if !globusClient.IsClientSet() {
		return ServerHandler{}, fmt.Errorf("AUTH error: Client is nil")
	}

	srcGroupTemplate, err := template.New("source group template").Parse(srcGroupTemplateBody)
	if err != nil {
		return ServerHandler{}, err
	}

	dstGroupTemplate, err := template.New("destination group template").Parse(dstGroupTemplateBody)
	if err != nil {
		return ServerHandler{}, err
	}

	dstPathTemplate, err := NewDestinationTemplate(dstPathTemplateBody)
	if err != nil {
		return ServerHandler{}, err
	}

	return ServerHandler{
		scicatUrl:             scicatUrl,
		globusClient:          globusClient,
		facilityCollectionIDs: facilityCollectionIDs,
		srcGroupTemplate:      srcGroupTemplate,
		dstGroupTemplate:      dstGroupTemplate,
		dstPathTemplate:       dstPathTemplate,
		taskPool:              taskPool,
	}, err
}

func (s ServerHandler) PostTransferTask(ctx context.Context, request PostTransferTaskRequestObject) (PostTransferTaskResponseObject, error) {
	ginCtx := ctx.(*gin.Context)

	// check facility id's and fetch collection id's
	sourceCollectionID, ok := s.facilityCollectionIDs[request.Params.SourceFacility]
	if !ok {
		return PostTransferTask403JSONResponse{
			Message: getPointerOrNil("invalid source facility"),
		}, nil
	}
	destCollectionID, ok := s.facilityCollectionIDs[request.Params.DestFacility]
	if !ok {
		return PostTransferTask403JSONResponse{
			Message: getPointerOrNil("invalid destination facility"),
		}, nil
	}

	u, ok := ginCtx.Get("scicatUser")
	if !ok {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("no user was found"),
		}, nil
	}

	// fetch scicat user
	scicatUser, ok := u.(User)
	if !ok {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("invalid user in context"),
			Details: getPointerOrNil(fmt.Sprintf("type found: '%s'", reflect.TypeOf(u))),
		}, nil
	}

	// fetch related dataset
	datasetUrl, err := url.JoinPath(s.scicatUrl, "datasets", url.QueryEscape(request.Params.ScicatPid))
	if err != nil {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("couldn't create dataset request url"),
			Details: getPointerOrNil(err.Error()),
		}, nil
	}

	datasetReq, err := http.NewRequest("GET", datasetUrl, nil)
	if err != nil {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("couldn't generate dataset request"),
			Details: getPointerOrNil(err.Error()),
		}, nil
	}
	datasetReq.Header.Set("Authorization", "Bearer "+scicatUser.ScicatToken)

	datasetResp, err := http.DefaultClient.Do(datasetReq)
	if err != nil {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("couldn't send dataset request to scicat backend"),
			Details: getPointerOrNil(err.Error()),
		}, nil
	}
	defer datasetResp.Body.Close()

	if datasetResp.StatusCode != 200 {
		body, _ := io.ReadAll(datasetResp.Body)
		return PostTransferTask400JSONResponse{
			GeneralErrorResponseJSONResponse: GeneralErrorResponseJSONResponse{
				Message: getPointerOrNil("the dataset with the given pid does not exist or you don't have access rights to it"),
				Details: getPointerOrNil(fmt.Sprintf("response status '%d', body '%s'", datasetResp.StatusCode, string(body))),
			},
		}, nil
	}

	datasetRespBody, err := io.ReadAll(datasetResp.Body)
	if err != nil {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("failed to read response body"),
			Details: getPointerOrNil(err.Error()),
		}, nil
	}

	var dataset ScicatDataset
	err = json.Unmarshal(datasetRespBody, &dataset)
	if err != nil {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("failed to unmarshal response body"),
			Details: getPointerOrNil(err.Error()),
		}, nil
	}

	// check for required group memberships
	var srcGroupBuf bytes.Buffer
	err = s.srcGroupTemplate.Execute(&srcGroupBuf, GroupTemplateData{FacilityName: request.Params.SourceFacility})
	if err != nil {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("group templating failed with source facility"),
			Details: getPointerOrNil(err.Error()),
		}, nil
	}

	var dstGroupBuf bytes.Buffer
	err = s.dstGroupTemplate.Execute(&dstGroupBuf, GroupTemplateData{FacilityName: request.Params.DestFacility})
	if err != nil {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("group templating failed with destination facility"),
			Details: getPointerOrNil(err.Error()),
		}, nil
	}

	requiredGroups := []string{srcGroupBuf.String(), dstGroupBuf.String(), dataset.OwnerGroup}
	missingGroups := []string{}
	for _, group := range requiredGroups {
		if !slices.Contains(scicatUser.Profile.AccessGroups, group) {
			missingGroups = append(missingGroups, group)
		}
	}

	if len(missingGroups) > 0 {
		return PostTransferTask401JSONResponse{
			Message: getPointerOrNil("you don't have the required access groups to request this transfer"),
			Details: getPointerOrNil(fmt.Sprintf("missing groups: '%v'", missingGroups)),
		}, nil
	}

	// request the transfer
	params := destPathParams{
		DatasetFolder: path.Base(dataset.SourceFolder),
		SourceFolder:  dataset.SourceFolder,
		Pid:           request.Params.ScicatPid,
		PidShort:      path.Base(request.Params.ScicatPid),
		PidPrefix:     path.Dir(request.Params.ScicatPid),
		PidEncoded:    url.PathEscape(request.Params.ScicatPid),
		Username:      scicatUser.Profile.Username,
	}

	sourcePath := dataset.SourceFolder
	destPath, err := s.dstPathTemplate.Execute(params)
	if err != nil {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("couldn't template destination folder for the transfer"),
			Details: getPointerOrNil(err.Error()),
		}, nil
	}

	var result globus.TransferResult
	if request.Body == nil {
		return PostTransferTask400JSONResponse{
			GeneralErrorResponseJSONResponse: GeneralErrorResponseJSONResponse{
				Message: getPointerOrNil("no body was sent with the request"),
			},
		}, nil
	}

	if request.Body.FileList != nil {
		// use filelist
		paths := make([]string, len(*request.Body.FileList))
		isSymlinks := make([]bool, len(*request.Body.FileList))
		for i, file := range *request.Body.FileList {
			paths[i] = file.Path
			isSymlinks[i] = file.IsSymlink
		}
		result, err = s.globusClient.TransferFileList(sourceCollectionID, sourcePath, destCollectionID, destPath, paths, isSymlinks, false)
	} else {
		// sync folders through globus
		result, err = s.globusClient.TransferFolderSync(sourceCollectionID, sourcePath, destCollectionID, destPath, false)
	}

	if err != nil {
		return PostTransferTask400JSONResponse{
			GeneralErrorResponseJSONResponse: GeneralErrorResponseJSONResponse{
				Message: getPointerOrNil(fmt.Sprintf("transfer request failed: %s", err.Error())),
			},
		}, nil
	}

	s.taskPool.AddTransferTask(result.TaskId, request.Params.ScicatPid)

	// return response
	return PostTransferTask200JSONResponse{
		TaskId: getPointerOrNil(result.TaskId),
	}, nil
}

func getPointerOrNil[T comparable](v T) *T {
	var a T
	if a == v {
		return nil
	} else {
		return &v
	}
}
