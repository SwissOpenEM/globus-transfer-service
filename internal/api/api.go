package api

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"slices"
	"text/template"

	"github.com/SwissOpenEM/globus"
	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml openapi.yaml

type ServerHandler struct {
	globusClient          globus.GlobusClient
	facilityCollectionIDs map[string]string
	srcGroupTemplate      *template.Template
	dstGroupTemplate      *template.Template
}

var _ StrictServerInterface = ServerHandler{}

func NewServerHandler(clientID string, clientSecret string, scopes []string, facilityCollectionIDs map[string]string, srcGroupTemplate string, dstGroupTemplate string) (ServerHandler, error) {
	// create server with service client
	var err error
	globusClient, err := globus.AuthCreateServiceClient(context.Background(), clientID, clientSecret, scopes)
	if err != nil {
		return ServerHandler{}, err
	}

	if !globusClient.IsClientSet() {
		return ServerHandler{}, fmt.Errorf("AUTH error: Client is nil")
	}

	srcTemplate, err := template.New("source group template").Parse(srcGroupTemplate)
	if err != nil {
		return ServerHandler{}, err
	}

	dstTemplate, err := template.New("destination group template").Parse(dstGroupTemplate)
	if err != nil {
		return ServerHandler{}, err
	}

	return ServerHandler{
		globusClient:          globusClient,
		facilityCollectionIDs: facilityCollectionIDs,
		srcGroupTemplate:      srcTemplate,
		dstGroupTemplate:      dstTemplate,
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

	// check for required user access groups
	scicatUser, ok := u.(User)
	if !ok {
		return PostTransferTask500JSONResponse{
			Message: getPointerOrNil("invalid user in context"),
			Details: getPointerOrNil(fmt.Sprintf("type found: '%s'", reflect.TypeOf(u))),
		}, nil
	}

	var srcGroupBuf bytes.Buffer
	err := s.srcGroupTemplate.Execute(&srcGroupBuf, GroupTemplateData{FacilityName: request.Params.SourceFacility})
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

	requiredGroups := []string{srcGroupBuf.String(), dstGroupBuf.String()}
	missingGroups := []string{}
	for _, group := range requiredGroups {
		if !slices.Contains(scicatUser.Profile.AccessGroups, group) {
			missingGroups = append(missingGroups, group)
		}
	}

	if len(missingGroups) > 0 {
		return PostTransferTask401JSONResponse{
			GeneralErrorResponseJSONResponse: GeneralErrorResponseJSONResponse{
				Message: getPointerOrNil("you don't have the required access groups to request this transfer"),
				Details: getPointerOrNil(fmt.Sprintf("missing groups: '%v'", missingGroups)),
			},
		}, nil
	}

	// request the transfer
	sourcePath := request.Params.SourcePath
	destPath := "/" + request.Params.ScicatPid
	_ = destPath

	/*result, err := s.globusClient.TransferFileList(sourceCollectionID, sourcePath, destCollectionID, destPath, []string{}, []bool{}, false)
	if err != nil {
		return TransferPostJob400JSONResponse{
			Message: getPointerOrNil(fmt.Sprintf("transfer request failed: %s", err.Error())),
		}, nil
	}*/

	result, err := s.globusClient.TransferFolderSync(sourceCollectionID, sourcePath, destCollectionID, "/service_user/"+request.Params.ScicatPid, false)
	if err != nil {
		return PostTransferTask400JSONResponse{
			Message: getPointerOrNil(fmt.Sprintf("transfer request failed: %s", err.Error())),
		}, nil
	}

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
