package main

import (
	"context"
	"log"
	"os"

	"github.com/SwissOpenEM/globus"
	"github.com/SwissOpenEM/globus-transfer-service/internal/api"
	"github.com/SwissOpenEM/globus-transfer-service/internal/config"
	"github.com/SwissOpenEM/globus-transfer-service/internal/serviceuser"
	"github.com/SwissOpenEM/globus-transfer-service/internal/tasks"
)

func main() {
	globusClientId := os.Getenv("GLOBUS_CLIENT_ID")
	globusClientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")
	scicatServiceUserUsername := os.Getenv("SCICAT_SERVICE_USER_USERNAME")
	scicatServiceUserPassword := os.Getenv("SCICAT_SERVICE_USER_PASSWORD")

	conf, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("couldn't read config: %s\n", err.Error())
	}

	serviceUser, err := serviceuser.CreateServiceUser(conf.ScicatUrl, scicatServiceUserUsername, scicatServiceUserPassword)
	if err != nil {
		log.Fatalf("couldn't create service user: %s\n", err.Error())
	}

	globusClient, err := globus.AuthCreateServiceClient(context.Background(), globusClientId, globusClientSecret, conf.GlobusScopes)
	if err != nil {
		log.Fatalf("couldn't create globus client: %s\n", err.Error())
	}

	taskPool := tasks.CreateTaskPool(conf.ScicatUrl, globusClient, serviceUser, conf.Task.MaxConcurrency, conf.Task.QueueSize, conf.Task.PollInterval)

	err = tasks.RestoreGlobusTransferJobsFromScicat(conf.ScicatUrl, serviceUser, taskPool)
	if err != nil {
		log.Fatalf("couldn't resume unfinished jobs: %s\n", err.Error())
	}

	serverHandler, err := api.NewServerHandler(globusClient, conf.GlobusScopes, conf.ScicatUrl, serviceUser, conf.FacilityCollectionIDs, conf.FacilitySrcGroupTemplate, conf.FacilityDstGroupTemplate, conf.DstPathTemplate, taskPool)
	if err != nil {
		log.Fatal(err)
	}

	server, err := api.NewServer(&serverHandler, conf.Port, conf.ScicatUrl)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.ListenAndServe())
}
