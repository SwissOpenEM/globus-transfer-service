package api

import (
	"context"
	"fmt"

	"github.com/SwissOpenEM/globus"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../openapi.yaml

type Server struct {
	globusClient                    globus.GlobusClient
	facilityToGlobusCollectionIDMap map[string]string
}

var _ StrictServerInterface = Server{}

func NewServer(clientID string, clientSecret string, scopes []string) (Server, error) {
	// create server with service client
	var err error
	globusClient, err := globus.AuthCreateServiceClient(context.Background(), clientID, clientSecret, scopes)
	if err != nil {
		return Server{}, err
	}
	if !globusClient.IsClientSet() {
		return Server{}, fmt.Errorf("AUTH error: Client is nil")
	}
	return Server{
		globusClient: globusClient,
	}, err
}

func (s Server) TransferPostJob(ctx context.Context, request TransferPostJobRequestObject) (TransferPostJobResponseObject, error) {
	sourceCollectionID, ok := s.facilityToGlobusCollectionIDMap[request.Params.SourceFacility]
	if !ok {
		return TransferPostJob403JSONResponse{
			Message: getPointerOrNil("invalid source facility"),
		}, nil
	}
	destCollectionID, ok := s.facilityToGlobusCollectionIDMap[request.Params.DestFacility]
	if !ok {
		return TransferPostJob403JSONResponse{
			Message: getPointerOrNil("invalid destination facility"),
		}, nil
	}

	sourcePath := request.Params.SourcePath
	destPath := "/" + request.Params.ScicatPid

	result, err := s.globusClient.TransferFileList(sourceCollectionID, sourcePath, destCollectionID, destPath, []string{}, []bool{}, false)
	if err != nil {
		return TransferPostJob400JSONResponse{
			Message: getPointerOrNil(fmt.Sprintf("transfer request failed: %s", err.Error())),
		}, nil
	}

	return TransferPostJob200JSONResponse{
		JobId: getPointerOrNil(result.RequestId),
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
