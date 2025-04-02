package main

import (
	"context"
	"log"
	"os"

	"github.com/SwissOpenEM/globus"
	"github.com/SwissOpenEM/globus-transfer-service/internal/api"
	"github.com/SwissOpenEM/globus-transfer-service/internal/config"
	"github.com/SwissOpenEM/globus-transfer-service/internal/tasks"
)

func main() {
	globusClientId := os.Getenv("GLOBUS_CLIENT_ID")
	globusClientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	conf, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("couldn't read config: %s\n", err.Error())
	}

	globusClient, err := globus.AuthCreateServiceClient(context.Background(), globusClientId, globusClientSecret, conf.GlobusScopes)
	if err != nil {
		log.Fatalf("couldn't create globus client %s\n", err.Error())
	}

	taskPool := tasks.CreateTaskPool(conf.ScicatUrl, globusClient, conf.Task.MaxConcurrency, conf.Task.QueueSize, conf.Task.PollInterval)

	serverHandler, err := api.NewServerHandler(globusClient, conf.GlobusScopes, conf.ScicatUrl, conf.FacilityCollectionIDs, conf.FacilitySrcGroupTemplate, conf.FacilityDstGroupTemplate, conf.DstPathTemplate, taskPool)
	if err != nil {
		log.Fatal(err)
	}

	server, err := api.NewServer(&serverHandler, conf.Port, conf.ScicatUrl)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.ListenAndServe())
}
