package api

import (
	"context"
	"fmt"

	"github.com/SwissOpenEM/globus"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml openapi.yaml

type ServerHandler struct {
	globusClient          globus.GlobusClient
	facilityCollectionIDs map[string]string
}

var _ StrictServerInterface = ServerHandler{}

func NewServerHandler(clientID string, clientSecret string, scopes []string, facilityCollectionIDs map[string]string) (ServerHandler, error) {
	// create server with service client
	var err error
	globusClient, err := globus.AuthCreateServiceClient(context.Background(), clientID, clientSecret, scopes)
	if err != nil {
		return ServerHandler{}, err
	}
	if !globusClient.IsClientSet() {
		return ServerHandler{}, fmt.Errorf("AUTH error: Client is nil")
	}
	return ServerHandler{
		globusClient:          globusClient,
		facilityCollectionIDs: facilityCollectionIDs,
	}, err
}

func (s ServerHandler) PostTransferTask(ctx context.Context, request PostTransferTaskRequestObject) (PostTransferTaskResponseObject, error) {
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
