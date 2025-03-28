package main

import (
	"log"
	"os"

	"github.com/SwissOpenEM/globus-transfer-service/internal/api"
	"github.com/SwissOpenEM/globus-transfer-service/internal/config"
)

func main() {
	globusClientId := os.Getenv("GLOBUS_CLIENT_ID")
	globusClientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	conf, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("couldn't read config: %s", err.Error())
	}

	serverHandler, err := api.NewServerHandler(globusClientId, globusClientSecret, conf.GlobusScopes, conf.FacilityCollectionIDs, conf.FacilitySrcGroupTemplate, conf.FacilityDstGroupTemplate, conf.ScicatUrl)
	if err != nil {
		log.Fatal(err)
	}

	server, err := api.NewServer(&serverHandler, conf.Port, conf.ScicatUrl)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.ListenAndServe())
}
