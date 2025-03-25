package api

import (
	"embed"
	"log"
	"net"
	"net/http"

	"fmt"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var swaggerYAML embed.FS

func NewServer(api *ServerHandler, port uint) *http.Server {
	swagger, err := GetSwagger()
	if err != nil {
		log.Fatalf("Error loading swagger spec: %s", err)
	}

	// not sure if this is needed??
	swagger.Servers = nil

	r := gin.New()

	// TODO: get the openapi.yaml embedded somehow
	r.GET("/openapi.yaml", func(c *gin.Context) {
		http.FileServer(http.FS(swaggerYAML))
	})

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	RegisterHandlers(r, NewStrictHandler(api, []StrictMiddlewareFunc{}))

	return &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort("0.0.0.0", fmt.Sprint(port)),
	}
}
