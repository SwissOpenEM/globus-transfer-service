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
		log.Fatal(err)
	}

	server, _ := api.NewServerHandler(globusClientId, globusClientSecret, conf.GlobusScopes, conf.FacilityCollectionIDs)
	//log.Fatal(server)
	_ = server
}
